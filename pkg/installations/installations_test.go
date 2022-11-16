package installations

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	database "github.com/xmtp/example-notification-server-go/pkg/db"
	"github.com/xmtp/example-notification-server-go/pkg/interfaces"
	"github.com/xmtp/example-notification-server-go/pkg/logging"
	"github.com/xmtp/example-notification-server-go/test"
)

const INSTALLATION_ID = "foo"
const TOKEN = "bar"

func createService(db *bun.DB) interfaces.Installations {
	return NewInstallationsService(
		context.Background(),
		logging.CreateLogger("console", "info"),
		db,
	)
}

func buildInstallation(installationId string, kind database.DeliveryMechanismKind, token string) interfaces.Installation {
	return interfaces.Installation{
		Id: installationId,
		DeliveryMechanism: interfaces.DeliveryMechanism{
			Kind:  kind,
			Token: token,
		},
	}
}

func Test_Register(t *testing.T) {
	ctx := context.Background()
	db, cleanup := test.CreateTestDb()
	defer cleanup()

	svc := createService(db)
	res, err := svc.Register(ctx, buildInstallation(INSTALLATION_ID, database.APNS, TOKEN))

	require.NoError(t, err)
	require.Equal(t, INSTALLATION_ID, res.InstallationId)

	installation := new(database.Installation)
	err = db.NewSelect().Model(installation).Relation("DeliveryMechanisms").Where("id = ?", INSTALLATION_ID).Scan(ctx)

	require.NoError(t, err)
	require.Equal(t, installation.Id, INSTALLATION_ID)
	require.Len(t, installation.DeliveryMechanisms, 1)
	require.Equal(t, installation.DeliveryMechanisms[0].Kind, database.APNS)
	require.Equal(t, installation.DeliveryMechanisms[0].Token, TOKEN)
}

func Test_RegisterDuplicate(t *testing.T) {
	var err error
	ctx := context.Background()
	db, cleanup := test.CreateTestDb()
	defer cleanup()

	svc := createService(db)

	req := buildInstallation(INSTALLATION_ID, database.APNS, TOKEN)
	_, err = svc.Register(ctx, req)

	require.NoError(t, err)

	firstInstallation := new(database.Installation)
	err = db.NewSelect().
		Model(firstInstallation).
		Relation("DeliveryMechanisms").
		Where("id = ?", INSTALLATION_ID).
		Scan(ctx)

	require.NoError(t, err)

	_, err = svc.Register(ctx, req)

	require.NoError(t, err)

	secondInstallation := new(database.Installation)
	err = db.NewSelect().
		Model(secondInstallation).
		Relation("DeliveryMechanisms").
		Where("id = ?", INSTALLATION_ID).
		Scan(ctx)

	require.NoError(t, err)
	require.True(t, firstInstallation.CreatedAt.Equal(secondInstallation.CreatedAt))
	require.Len(t, secondInstallation.DeliveryMechanisms, 1)
	require.Equal(t, firstInstallation.DeliveryMechanisms[0].CreatedAt, secondInstallation.DeliveryMechanisms[0].CreatedAt)
	require.NotEqual(t, firstInstallation.DeliveryMechanisms[0].UpdatedAt, secondInstallation.DeliveryMechanisms[0].UpdatedAt)

}

func Test_RegisterUpdate(t *testing.T) {
	ctx := context.Background()
	db, cleanup := test.CreateTestDb()
	defer cleanup()

	var err error
	svc := createService(db)

	token1 := "token1"
	token2 := "token2"

	req1 := buildInstallation(INSTALLATION_ID, database.APNS, token1)

	_, err = svc.Register(ctx, req1)
	require.NoError(t, err)

	firstInstallation := new(database.Installation)
	err = db.NewSelect().
		Model(firstInstallation).
		Relation("DeliveryMechanisms").
		Where("id = ?", INSTALLATION_ID).
		Scan(ctx)

	require.NoError(t, err)

	req2 := buildInstallation(INSTALLATION_ID, database.APNS, token2)
	_, err = svc.Register(ctx, req2)

	require.NoError(t, err)

	secondInstallation := new(database.Installation)
	err = db.NewSelect().
		Model(secondInstallation).
		Relation("DeliveryMechanisms").
		Where("id = ?", INSTALLATION_ID).
		Scan(ctx)

	require.NoError(t, err)
	require.Len(t, secondInstallation.DeliveryMechanisms, 2)
	require.Equal(t, secondInstallation.DeliveryMechanisms[1].Token, token2)
	require.True(t, firstInstallation.CreatedAt.Equal(secondInstallation.CreatedAt))
	require.NotEqual(t, firstInstallation.DeliveryMechanisms[0].CreatedAt, secondInstallation.DeliveryMechanisms[1].CreatedAt)
}

func Test_Delete(t *testing.T) {
	ctx := context.Background()
	db, cleanup := test.CreateTestDb()
	defer cleanup()
	svc := createService(db)

	createReq := buildInstallation(INSTALLATION_ID, database.APNS, TOKEN)
	_, err := svc.Register(ctx, createReq)

	require.NoError(t, err)

	err = svc.Delete(ctx, INSTALLATION_ID)
	require.NoError(t, err)

	install := new(database.Installation)
	err = db.NewSelect().
		Model(install).
		Where("id = ?", INSTALLATION_ID).
		Scan(ctx)

	require.NoError(t, err)
	require.NotNil(t, install.DeletedAt)
}

func Test_Get(t *testing.T) {
	ctx := context.Background()
	db, _ := test.CreateTestDb()
	// defer cleanup()
	svc := createService(db)

	installationIds := []string{"install1", "install2", "install3"}
	for _, installationId := range installationIds {
		_, err := svc.Register(ctx, buildInstallation(installationId, database.APNS, TOKEN))
		require.NoError(t, err)
	}

	installs, err := svc.GetInstallations(ctx, installationIds)
	require.NoError(t, err)
	require.Len(t, installs, len(installationIds))
	for i, install := range installs {
		require.Equal(t, install.Id, installationIds[len(installationIds)-i-1])
		require.Equal(t, install.DeliveryMechanism.Token, TOKEN)
	}
}

func Test_GetMultiple(t *testing.T) {
	ctx := context.Background()
	db, cleanup := test.CreateTestDb()
	defer cleanup()
	svc := createService(db)

	tokens := []string{"token1", "token2", "token3"}
	for _, token := range tokens {
		_, err := svc.Register(ctx, buildInstallation(INSTALLATION_ID, database.APNS, token))
		require.NoError(t, err)
	}

	res, err := svc.GetInstallations(ctx, []string{INSTALLATION_ID})
	require.NoError(t, err)
	require.Len(t, res, 1)
	require.Equal(t, res[0].DeliveryMechanism.Token, "token3")
}
