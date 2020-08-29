package syncer_test

import (
	"encoding/json"
	"log"
	"os"
	"strconv"
	"time"

	"git.topvdn.com/web/goantelope/syncer"
)

// 从环境变量获取相关配置
var (
	lyID  = os.Getenv("LYID")
	lyKey = os.Getenv("LYKEY")

	webhookSyncURL      = os.Getenv("WEBHOOK_SYNC_URL")
	webhookSyncerName   = "webhook-example-syncer"
	webhookSyncInterval = 1

	testAppID    = os.Getenv("WEBHOOK_TEST_APPID")
	testCIDStr   = os.Getenv("WEBHOOK_TEST_CID")
	testHookType = os.Getenv("WEBHOOK_TEST_HOOKTYPE")
)

func ExampleWebhookSyncer() {
	webhookSyncer := syncer.NewWebhookSyncer(
		webhookSyncerName, lyID, lyKey, webhookSyncURL, webhookSyncInterval)
	webhookSyncer.Run()

	for {
		log.Println("webhooks", webhookSyncer.Webhooks)
		data, err := json.Marshal(webhookSyncer)
		if err != nil {
			panic(err)
		}
		log.Println("data", string(data))

		cid, err := strconv.ParseUint(testCIDStr, 10, 32)
		if err != nil {
			panic(err)
		}

		webhooks := webhookSyncer.GetWebhooks(uint32(cid), testAppID, testHookType)
		log.Printf(
			"cid: %v, appid: %v, webhook type: %v\n",
			testCIDStr, testAppID, testHookType)
		log.Println("webhooks:")
		for _, webhook := range webhooks {
			webhookJSON, err := json.Marshal(webhook)
			if err != nil {
				panic(err)
			}
			log.Println(string(webhookJSON))
		}
		time.Sleep(time.Second)
	}
	// Output:
}
