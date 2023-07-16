package handlers

import (
	"context"
	"fmt"
	"os"
	"strings"

	"inventory-manager/initialization"
	"inventory-manager/services/ocr"

	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

type PicAction struct { /*图片*/
}

func (*PicAction) Execute(a *ActionInfo) bool {

	if a.info.msgType == "image" {
		//保存图片
		imageKey := a.info.imageKey
		// fmt.Printf("fileKey: %s \n", imageKey)
		msgId := a.info.msgId
		//fmt.Println("msgId: ", *msgId)
		req := larkim.NewGetMessageResourceReqBuilder().MessageId(
			*msgId).FileKey(imageKey).Type("image").Build()
		resp, err := initialization.GetLarkClient().Im.MessageResource.Get(context.Background(), req)
		//fmt.Println(resp, err)
		if err != nil {
			//fmt.Println(err)
			replyMsg(*a.ctx, fmt.Sprintf("🤖️：图片下载失败，请稍后再试～\n 错误信息: %v", err),
				a.info.msgId)
			return false
		}

		f := fmt.Sprintf("%s.png", imageKey)
		resp.WriteFile(f)
		defer os.Remove(f)

		content, err := ocr.RecognizeTableOcr(a.handler.config, f)
		if err != nil {
			replyMsg(*a.ctx, fmt.Sprintf("🤖️：无法解析图片，请发送原图并尝试重新操作～"),
				a.info.msgId)
			return false
		}
		a.info.qParsed = strings.Trim(content, " ")
	}

	return true
}
