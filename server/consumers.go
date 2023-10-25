package server

import (
	"encoding/json"
	"github.com/ap-pauloafonso/investor-chat/queue"
	"github.com/ap-pauloafonso/investor-chat/utils"
	"github.com/ap-pauloafonso/investor-chat/websocket"
)

func (s *Server) initConsumers() {
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

	err = s.q.ConsumeUpdateChannelsCommand(func(payload []byte) error {
		// get all the channelConnections
		channels, err := s.channelService.GetAllChannels()
		if err != nil {
			return err
		}

		return s.webSocketHandler.HandleChannelsUpdate(channels)
	})
	if err != nil {
		utils.LogErrorFatal(err)
	}

	err = s.q.ConsumeBotCommandResponse(func(payload []byte) error {

		var obj queue.BotCommandResponse
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
