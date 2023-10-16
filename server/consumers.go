package server

import (
	"encoding/json"
	"investorchat/queue"
	"investorchat/utils"
	"investorchat/websocket"
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
		channels, err := s.chatService.GetAllChannels()
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

	err = s.q.ConsumeUserMessageCommandForStorage(func(payload []byte) error {
		var obj websocket.MessageObj
		err := json.Unmarshal(payload, &obj)
		if err != nil {
			return err
		}

		err = s.chatService.SaveMessage(obj.Channel, obj.Username, obj.Message, obj.Time)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		utils.LogErrorFatal(err)
	}
}