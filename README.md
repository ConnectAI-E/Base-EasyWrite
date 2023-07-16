# Base-EasyWrite
🪖 实验项目 希望通过语音或票据信息，方便地向多维表格录入数据。

## 真实场景
客户每天有海量的餐饮物资需要入库，但是手动的填写多维表格又非常麻烦

 (语音识别+图片OCR) x OpenAI的信息摘要能力 = 多模态的数据录入

## 最终效果

https://github.com/ConnectAI-E/Base-EasyWrite/assets/50035229/2c1479d7-49c0-4c75-9bc9-14cece4bd13a


## 项目核心的Prompt

ocr识别后
```
prompt := `下面是一些通过 OCR 识别图片生成的文字 【` + a.info.qParsed + `】，请整理返回 存货名称/品名/数量/金额/日期`
```

语音识别后
```
prompt := `你是仓库管理员,产品数据库有如下商品：【` + productNames + `】。供应商有：【` + suppliers + `】，你可以将下面的文本转换为json数组的入库记录单。
	下面是部分实例
	---
	Q: 【李峰供应商，送来吊龙三斤，总价143元】
	A: [{
			"date":` + today + `,
			"productName":"吊龙",
			"companyName":"李峰",
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
	Q: 【线下自购莲藕100斤 合计189元】
	A: [{
			"date":` + today + `,
			"productName":"藕",
			"companyName":"线下自购",
			"totalNumber":100,
			"totalPrice":189,
		}]
	Q: 【6月3日，供应商小军收入豆奶10瓶子 合计100元】
	A: [{
			"date":"2023/06/03",
			"productName":"唯怡豆奶（塑框玻璃瓶）",
			"companyName":"小军",
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
```

