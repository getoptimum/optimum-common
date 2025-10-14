package entities_test

import (
	"encoding/json"
	"testing"

	"github.com/getoptimum/optimum-common/pkg/entities"
	"github.com/getoptimum/optimum-common/pkg/utils"
	"github.com/stretchr/testify/require"
)

func TestMarshalUnMarshalP2PMessage(t *testing.T) {
	t.Run("valid_message", func(t *testing.T) {
		validMsg := &entities.P2PMessage{
			SourceNodeID:   "sourceNodeID",
			UpstreamPeerID: "upstreamPeerID",
			Topic:          "testTopic",
			Message:        []byte("testMessage"),
		}
		data, err := validMsg.Marshal()
		require.NoError(t, err)
		require.NotEmpty(t, data)

		msg, err := entities.UnmarshalP2PMessage(data)
		require.NoError(t, err)
		require.NotNil(t, msg)
		require.Equal(t, validMsg.SourceNodeID, msg.SourceNodeID)
		require.Equal(t, validMsg.UpstreamPeerID, msg.UpstreamPeerID)
		require.Equal(t, validMsg.Topic, msg.Topic)
		require.Equal(t, validMsg.Message, msg.Message)
	})
	t.Run("invalid_message", func(t *testing.T) {
		wrongMessage := map[string]string{
			"invalid_field": "invalid_value",
			"another_field": "another_value",
		}
		data, err := json.Marshal(wrongMessage)
		require.NoError(t, err)
		require.NotEmpty(t, data)

		msg, err := entities.UnmarshalP2PMessage(data)
		require.Error(t, err)
		require.Nil(t, msg)
	})
}

func TestMessageIDFn(t *testing.T) {
	t.Run("prefilled_message_id", func(t *testing.T) {
		validMsg := &entities.P2PMessage{
			SourceNodeID:   "sourceNodeID",
			UpstreamPeerID: "upstreamPeerID",
			Topic:          "testTopic",
			Message:        []byte("testMessage"),
		}
		data, err := validMsg.Marshal()
		require.NoError(t, err)

		require.Equal(t, utils.HashSHA256([]byte("testMessage")), entities.MessageIDFn(data))
	})
	t.Run("empty_message_id", func(t *testing.T) {
		wrongMessage := map[string]string{
			"invalid_field": "invalid_value",
			"another_field": "another_value",
		}
		data, err := json.Marshal(wrongMessage)
		require.NoError(t, err)

		require.Equal(t, utils.HashSHA256(data), entities.MessageIDFn(data))
	})
}
