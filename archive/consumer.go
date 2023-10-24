package archive

import (
	"encoding/json"
	"investorchat/utils"
	"investorchat/websocket"
)

func (s *Service) InitConsumer() {
	err := s.q.ConsumeUserMessageCommandForStorage(func(payload []byte) error {
		var obj websocket.MessageObj
		err := json.Unmarshal(payload, &obj)
		if err != nil {
			return err
		}

		err = s.SaveMessage(obj.Channel, obj.Username, obj.Message, obj.Time)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		utils.LogErrorFatal(err)
	}
}
