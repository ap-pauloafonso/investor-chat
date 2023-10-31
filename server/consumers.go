package server

import (
	"context"
	"encoding/json"
	"github.com/ap-pauloafonso/investor-chat/eventbus"
	"github.com/ap-pauloafonso/investor-chat/utils"
	"github.com/ap-pauloafonso/investor-chat/websocket"
)

func (s *Server) initConsumers(ctx context.Context) {
	err := s.q.ConsumeUserMessageCommandForWSBroadcast(func(payload []byte) error {

		var obj websocket.MessageObj
		err := json.Unmarshal(payload, &obj)
		if err != nil {
			return err
		}

		return s.webSocketHandler.BroadcastMessage(obj.Username, obj.Channel, obj.Message, false, obj.Time)

	})
	if err != nil {
		utils.LogErrorFatal(err)
	}

	err = s.q.ConsumeUpdateChannelsCommand(ctx, func(ctx context.Context, payload []byte) error {
		// get all the channelConnections
		channels, err := s.channelService.GetAllChannels(ctx)
		if err != nil {
			return err
		}

		return s.webSocketHandler.HandleChannelsUpdate(ctx, channels)
	})
	if err != nil {
		utils.LogErrorFatal(err)
	}

	err = s.q.ConsumeBotCommandResponse(func(payload []byte) error {

		var obj eventbus.BotCommandResponse
		err := json.Unmarshal(payload, &obj)
		if err != nil {
			return err
		}

		return s.webSocketHandler.BroadcastMessage("BOT", obj.Channel, obj.GeneratedMessage, true, obj.Time)

	})
	if err != nil {
		utils.LogErrorFatal(err)
	}

}
