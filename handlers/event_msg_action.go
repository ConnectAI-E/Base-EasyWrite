package handlers

import (
	"encoding/json"
	"fmt"
	"inventory-manager/services/lark"
	"inventory-manager/services/openai"
	"log"
	"strconv"
	"strings"
	"time"
)

type MessageAction struct { /*消息*/
}

func (*MessageAction) Execute(a *ActionInfo) bool {
	var err error
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("%v", e)
		}
		if err != nil {
			fmt.Println("err... ", err)
			SendErrorWriteCard(*a.ctx, a.info.msgId, err)
		}
	}()

	var msgContent string
	if a.info.msgType == "image" {
		var err error
		msgContent, err = getPromptContentByOcrResult(a)
		if err != nil {
			SendErrorWriteCard(*a.ctx, a.info.msgId, err)
			return false
		}
	} else {
		msgContent = a.info.qParsed
	}

	products := a.handler.productCache.GetProducts()
	if products == nil {
		var err error
		products, err = lark.GetAllProducts(a.handler.config)
		if err != nil {
			SendErrorWriteCard(*a.ctx, a.info.msgId, err)
			return false
		}
		a.handler.productCache.Clear()
		a.handler.productCache.SetProducts(products)
	}
	productNameStr := ""
	for name := range products {
		productNameStr += name + ","
	}

	suppliers := a.handler.supplierCache.GetSuppliers()
	if len(suppliers) == 0 {
		var err error
		suppliers, err = lark.GetAllSuppliers(a.handler.config)
		if err != nil {
			SendErrorWriteCard(*a.ctx, a.info.msgId, err)
			return false
		}
		a.handler.supplierCache.Clear()
		a.handler.supplierCache.SetSuppliers(suppliers)
	}
	allSuppliers := ""
	for name := range suppliers {
		allSuppliers += name
	}

	prompt := getPrompt(msgContent, a.info.msgType, productNameStr, allSuppliers)

	var msg []openai.Messages
	msg = append(msg, openai.Messages{
		Role: "user", Content: prompt,
	})
	// 精准的任务，temperature设置为0.0 效果更好
	completions, err := a.handler.gpt.Completions(msg, 0)
	log.Println("msgId", *a.info.msgId, "completions.Content", completions.Content)
	if err != nil {
		SendErrorWriteCard(*a.ctx, a.info.msgId, err)
		return false
	}

	var recordList []*lark.ReceiptRecord
	content := strings.TrimPrefix(completions.Content, "A: ")
	err = json.Unmarshal([]byte(content), &recordList)
	if err != nil {
		SendErrorWriteCard(*a.ctx, a.info.msgId, err)
		return false
	}
	result, firstLink, err := lark.CreateTableRecord(a.handler.config, recordList, products, suppliers)
	if err != nil {
		SendErrorWriteCard(*a.ctx, a.info.msgId, err)
		return false
	}

	SendSuccessWriteCard(*a.ctx, a.info.msgId, result, firstLink)

	return true
}

// if len(recordList) == 1 {
// 	isRelevant, err := checkIfRelevant(a, a.info.qParsed, recordList[0].ProductName,productNameStr)
// 	if err != nil {
// 		SendErrorWriteCard(*a.ctx, a.info.msgId, err)
// 		return false
// 	}
// 	if !isRelevant {
// 		SendErrorWriteCard(*a.ctx, a.info.msgId, nil)
// 		return false
// 	}
// }
// func checkIfRelevant(a *ActionInfo, msgContent, productName string, allProductNames string) (bool, error) {
// 	prompt := `第一个输入信息是：【` + msgContent + `】，第二个输入信息是一组商品名：【` + allProductNames + `】，输出的信息是：【` + productName + `】，
// 	请问输出是否和第一个输入信息相关联，并且输出信息是从第二个输入信息中挑选的商品名，如果是请返回 true，其余情况均返回 false。
// 	直接返回 true 或者 false，不需要其他提示信息。`
// 	prompt = strings.ReplaceAll(strings.Trim(prompt, " "), "\n", "")

// 	fmt.Println(prompt)
// 	var msg []openai.Messages
// 	msg = append(msg, openai.Messages{
// 		Role: "user", Content: prompt,
// 	})
// 	completions, err := a.handler.gpt.Completions(msg, 0)
// 	if err != nil {
// 		return false, err
// 	}
// 	fmt.Println(completions.Content)
// 	isRelevant, err := strconv.ParseBool(completions.Content)
// 	if err != nil {
// 		return false, err
// 	}
// 	return isRelevant, nil
// }

func getPromptContentByOcrResult(a *ActionInfo) (string, error) {
	prompt := `下面是一些通过 OCR 识别图片生成的文字 【` + a.info.qParsed + `】，请整理返回 存货名称/品名/数量/金额/日期`
	prompt = strings.ReplaceAll(strings.Trim(prompt, " "), "\n", "")

	var msg []openai.Messages
	msg = append(msg, openai.Messages{
		Role: "user", Content: prompt,
	})
	completions, err := a.handler.gpt.Completions(msg, 0)
	if err != nil {
		return "", err
	}

	return completions.Content, nil
}

func getPrompt(content, msgType, productNames, suppliers string) string {
	location, _ := time.LoadLocation("Asia/Shanghai")
	now := time.Now().In(location)
	year := now.Year()
	today := now.Format("2006/01/02")

	prompt := `你是仓库管理员,产品数据库有如下商品：【` + productNames + `】。供应商有：【` + suppliers + `】，你可以将下面的文本转换为json数组的入库记录单。
	下面是部分实例
	---
	Q: 【王峰供应商，送来吊龙三斤，总价143元】
	A: [{
			"date":` + today + `,
			"productName":"吊龙",
			"companyName":"王峰",
			"totalNumber":3,
			"totalPrice":143,
		}]
	Q:【3月23号，生菜4.2斤，12块6，平菇，1.1斤，5块5】
	A: [{
			"date":"2023/03/23",
			"productName": "生菜",
			"companyName":"",
			"totalNumber": 4.2,
			"totalPrice": 12.6
		},
		{
			"date":"2023/03/23",
			"productName": "平菇",
			"companyName":"",
			"totalNumber": 1.1,
			"totalPrice": 5.5
		}]
	Q: 【线下自购莲藕100斤 合计89元】
	A: [{
			"date":` + today + `,
			"productName":"藕",
			"companyName":"线下自购",
			"totalNumber":100,
			"totalPrice":89,
		}]
	Q: 【6月3日，供应商高军收入豆奶10瓶子 合计100元】
	A: [{
			"date":"2023/06/03",
			"productName":"唯怡豆奶（塑框玻璃瓶）",
			"companyName":"高军",
			"totalNumber":10,
			"totalPrice"'':100,
		}]
	----
	那么，Q:【` + content + `】 转换成json数组的入库单结果是什么？直接返回 json数组的代码，不要其他提示信息。 
	请注意：
	1. 如果没说明时间，那么 date 字段就是 UTC+8 的现在日期；
	2. 如果没说明年份，默认是` + strconv.Itoa(year) + `年，比如 3月23号，date 就是` + strconv.Itoa(year) + `/03/23；
	3. productName 根据产品数据库中的名称，对提供的信息进行转换，结果请从产品数据库中选择，比如提供的信息是 雪花超级勇闯(蓝Super X)，产品数据库中
	是 雪花超级勇闯（蓝SuperX），那么请返回 雪花超级勇闯（蓝SuperX）；
	4. companyName 请从供应商中选择，如果没有匹配的，companyName 为空字符串；
	5. 可能会出现部分错别字，比如把 “5斤” 写成 “5金”。`

	return strings.ReplaceAll(strings.Trim(prompt, " "), "\n", "")
}
