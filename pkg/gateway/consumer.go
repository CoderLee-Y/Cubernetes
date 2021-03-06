package gateway

import (
	"Cubernetes/pkg/gateway/types"
	"context"
	"encoding/json"
	"log"
)

func (rg *RuntimeGateway) ListenReturnTopic() {
	ctx := context.Background()
	for {
		msgByte, err := rg.reader.ReadMessage(ctx)
		if err != nil {
			log.Printf("[Error]: read message failed")
			continue
		}

		msg := types.MQMessageResponse{}
		err = json.Unmarshal(msgByte.Value, &msg)
		if err != nil {
			log.Printf("[Error]: parse return message failed, msg: %v", string(msgByte.Value))
			return
		}

		rg.mapMutex.Lock()
		channel, ok := rg.channelMap[msg.RequestUID]
		delete(rg.channelMap, msg.RequestUID)
		rg.mapMutex.Unlock()

		if ok {
			log.Printf("[INFO]: Get resp from MQ, UID is %v", msg.RequestUID)
			channel <- msg
		} else {
			log.Printf("[Error]: return a resp and not found in channel map")
		}
	}
}
