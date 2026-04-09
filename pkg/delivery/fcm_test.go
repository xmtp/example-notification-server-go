package delivery

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xmtp/example-notification-server-go/pkg/interfaces"
)

func TestFcm_DataIncludesPayloadFormat(t *testing.T) {
	req := buildDeliveryRequest(t, interfaces.PayloadFormatV3)

	data := buildFcmData(req)
	require.Equal(t, "v3", data["payloadFormat"])
}

func Test_BuildFcmData_TopicField(t *testing.T) {
	req := buildDeliveryRequest(t, interfaces.PayloadFormatV3)

	data := buildFcmData(req)
	require.Equal(t, deliveryTestTopic, data["topic"])
	require.NotContains(t, data, "topicBytesB64")
	require.Equal(t, base64.StdEncoding.EncodeToString([]byte("test")), data["encryptedMessage"])
	require.Equal(t, "v3-conversation", data["messageType"])
}

func Test_BuildFcmData_V4TopicBytesB64(t *testing.T) {
	req := buildDeliveryRequest(t, interfaces.PayloadFormatV4)

	data := buildFcmData(req)
	require.Equal(t, deliveryTestTopic, data["topic"])
	require.Equal(t, req.TopicBytesB64, data["topicBytesB64"])
	require.Equal(t, "v4", data["payloadFormat"])
}
