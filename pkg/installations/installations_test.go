package installations

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/xmtp/example-notification-server-go/pkg/interfaces"
	"github.com/xmtp/example-notification-server-go/pkg/subscriptions"
	"github.com/xmtp/example-notification-server-go/pkg/testutils"
)

const INSTALLATION_ID = "foo"
const TOKEN = "bar"

type storedInstallation struct {
	ID                 string
	CreatedAt          time.Time
	DeletedAt          sql.NullTime
	DeliveryMechanisms []storedDeliveryMechanism
}

type storedDeliveryMechanism struct {
	ID        int64
	CreatedAt time.Time
	UpdatedAt time.Time
	Kind      string
	Token     string
}

func createService(t *testing.T, db *sql.DB) interfaces.Installations {
	t.Helper()
	return NewInstallationsService(
		testutils.TestLogger(t),
		db,
	)
}

func buildInstallation(installationID string, kind interfaces.DeliveryMechanismKind, token string) interfaces.Installation {
	return interfaces.Installation{
		Id: installationID,
		DeliveryMechanism: interfaces.DeliveryMechanism{
			Kind:  kind,
			Token: token,
		},
	}
}

func fetchInstallation(t *testing.T, ctx context.Context, db *sql.DB, installationID string) storedInstallation {
	t.Helper()

	var result storedInstallation
	err := db.QueryRowContext(
		ctx,
		`SELECT id, created_at, deleted_at FROM installations WHERE id = $1`,
		installationID,
	).Scan(&result.ID, &result.CreatedAt, &result.DeletedAt)
	require.NoError(t, err)

	rows, err := db.QueryContext(
		ctx,
		`SELECT id, created_at, updated_at, kind, token
		 FROM device_delivery_mechanisms
		 WHERE installation_id = $1
		 ORDER BY id ASC`,
		installationID,
	)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, rows.Close())
	}()

	for rows.Next() {
		var mechanism storedDeliveryMechanism
		require.NoError(t, rows.Scan(
			&mechanism.ID,
			&mechanism.CreatedAt,
			&mechanism.UpdatedAt,
			&mechanism.Kind,
			&mechanism.Token,
		))
		result.DeliveryMechanisms = append(result.DeliveryMechanisms, mechanism)
	}
	require.NoError(t, rows.Err())

	return result
}

func Test_Register(t *testing.T) {
	ctx := t.Context()
	db := testutils.CreateTestDb(t)

	svc := createService(t, db)
	res, err := svc.Register(ctx, buildInstallation(INSTALLATION_ID, interfaces.APNS, TOKEN))

	require.NoError(t, err)
	require.Equal(t, INSTALLATION_ID, res.InstallationId)

	installation := fetchInstallation(t, ctx, db, INSTALLATION_ID)
	require.Equal(t, INSTALLATION_ID, installation.ID)
	require.Len(t, installation.DeliveryMechanisms, 1)
	require.Equal(t, string(interfaces.APNS), installation.DeliveryMechanisms[0].Kind)
	require.Equal(t, TOKEN, installation.DeliveryMechanisms[0].Token)
}

func Test_RegisterDuplicate(t *testing.T) {
	ctx := t.Context()
	db := testutils.CreateTestDb(t)
	svc := createService(t, db)

	req := buildInstallation(INSTALLATION_ID, interfaces.APNS, TOKEN)
	_, err := svc.Register(ctx, req)
	require.NoError(t, err)

	first := fetchInstallation(t, ctx, db, INSTALLATION_ID)

	_, err = svc.Register(ctx, req)
	require.NoError(t, err)

	second := fetchInstallation(t, ctx, db, INSTALLATION_ID)
	require.True(t, first.CreatedAt.Equal(second.CreatedAt))
	require.Len(t, second.DeliveryMechanisms, 1)
	require.Equal(t, first.DeliveryMechanisms[0].CreatedAt, second.DeliveryMechanisms[0].CreatedAt)
	require.NotEqual(t, first.DeliveryMechanisms[0].UpdatedAt, second.DeliveryMechanisms[0].UpdatedAt)
}

func Test_RegisterUpdate(t *testing.T) {
	ctx := t.Context()
	db := testutils.CreateTestDb(t)
	svc := createService(t, db)

	req1 := buildInstallation(INSTALLATION_ID, interfaces.APNS, "token1")
	_, err := svc.Register(ctx, req1)
	require.NoError(t, err)

	first := fetchInstallation(t, ctx, db, INSTALLATION_ID)

	req2 := buildInstallation(INSTALLATION_ID, interfaces.APNS, "token2")
	_, err = svc.Register(ctx, req2)
	require.NoError(t, err)

	second := fetchInstallation(t, ctx, db, INSTALLATION_ID)
	require.Len(t, second.DeliveryMechanisms, 2)
	require.Equal(t, "token2", second.DeliveryMechanisms[1].Token)
	require.True(t, first.CreatedAt.Equal(second.CreatedAt))
	require.NotEqual(t, first.DeliveryMechanisms[0].CreatedAt, second.DeliveryMechanisms[1].CreatedAt)
}

func Test_Delete(t *testing.T) {
	ctx := t.Context()
	db := testutils.CreateTestDb(t)
	svc := createService(t, db)

	_, err := svc.Register(ctx, buildInstallation(INSTALLATION_ID, interfaces.APNS, TOKEN))
	require.NoError(t, err)

	err = svc.Delete(ctx, INSTALLATION_ID)
	require.NoError(t, err)

	installation := fetchInstallation(t, ctx, db, INSTALLATION_ID)
	require.True(t, installation.DeletedAt.Valid)
}

func Test_DeleteAndRegisterAgain(t *testing.T) {
	ctx := t.Context()
	db := testutils.CreateTestDb(t)
	svc := createService(t, db)

	_, err := svc.Register(ctx, buildInstallation(INSTALLATION_ID, interfaces.APNS, TOKEN))
	require.NoError(t, err)

	err = svc.Delete(ctx, INSTALLATION_ID)
	require.NoError(t, err)

	_, err = svc.Register(ctx, buildInstallation(INSTALLATION_ID, interfaces.APNS, TOKEN))
	require.NoError(t, err)

	installation := fetchInstallation(t, ctx, db, INSTALLATION_ID)
	require.False(t, installation.DeletedAt.Valid)
}

func Test_Get(t *testing.T) {
	ctx := t.Context()
	db := testutils.CreateTestDb(t)
	svc := createService(t, db)

	installationIDs := []string{"install1", "install2", "install3"}
	for _, installationID := range installationIDs {
		_, err := svc.Register(ctx, buildInstallation(installationID, interfaces.APNS, TOKEN))
		require.NoError(t, err)
	}

	installs, err := svc.GetInstallations(ctx, installationIDs)
	require.NoError(t, err)
	require.Len(t, installs, len(installationIDs))

	for i, install := range installs {
		require.Equal(t, installationIDs[len(installationIDs)-i-1], install.Id)
		require.Equal(t, TOKEN, install.DeliveryMechanism.Token)
	}
}

func Test_GetMultiple(t *testing.T) {
	ctx := t.Context()
	db := testutils.CreateTestDb(t)
	svc := createService(t, db)

	for _, token := range []string{"token1", "token2", "token3"} {
		_, err := svc.Register(ctx, buildInstallation(INSTALLATION_ID, interfaces.APNS, token))
		require.NoError(t, err)
	}

	res, err := svc.GetInstallations(ctx, []string{INSTALLATION_ID})
	require.NoError(t, err)
	require.Len(t, res, 1)
	require.Equal(t, "token3", res[0].DeliveryMechanism.Token)
}

func Test_DeleteDeactivatesSubscriptions(t *testing.T) {
	ctx := t.Context()
	db := testutils.CreateTestDb(t)

	installSvc := createService(t, db)
	subSvc := subscriptions.NewSubscriptionsService(
		testutils.TestLogger(t), db,
	)

	_, err := installSvc.Register(ctx, buildInstallation(INSTALLATION_ID, interfaces.APNS, TOKEN))
	require.NoError(t, err)

	err = subSvc.Subscribe(ctx, INSTALLATION_ID, []string{"topic1", "topic2"})
	require.NoError(t, err)

	err = installSvc.Delete(ctx, INSTALLATION_ID)
	require.NoError(t, err)

	// Subscriptions should be deactivated
	subs, err := subSvc.GetSubscriptions(ctx, "topic1", 1)
	require.NoError(t, err)
	require.Len(t, subs, 0)
}

func Test_GetDeleted(t *testing.T) {
	ctx := t.Context()
	db := testutils.CreateTestDb(t)
	svc := createService(t, db)

	_, err := svc.Register(ctx, buildInstallation(INSTALLATION_ID, interfaces.APNS, TOKEN))
	require.NoError(t, err)

	err = svc.Delete(ctx, INSTALLATION_ID)
	require.NoError(t, err)

	results, err := svc.GetInstallations(ctx, []string{INSTALLATION_ID})
	require.NoError(t, err)
	require.Len(t, results, 0)
}
