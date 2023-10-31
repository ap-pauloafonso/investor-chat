package archive

import (
	"context"
	"encoding/json"
	"github.com/ap-pauloafonso/investor-chat/utils"
	"github.com/ap-pauloafonso/investor-chat/websocket"
)

func (s *Service) InitConsumer(ctx context.Context) {
	err := s.eventbus.ConsumeUserMessageCommandForStorage(func(payload []byte) error {
		var obj websocket.MessageObj
		err := json.Unmarshal(payload, &obj)
		if err != nil {
			return err
		}

		err = s.SaveMessage(ctx, obj.Channel, obj.Username, obj.Message, obj.Time)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		utils.LogErrorFatal(err)
	}
}
