# Base-EasyWrite
🪖 实验项目 希望通过语音或票据信息，方便地向多维表格录入数据。

## 最终效果

https://github.com/ConnectAI-E/Base-EasyWrite/assets/50035229/2c1479d7-49c0-4c75-9bc9-14cece4bd13a


## 真实场景+AI思路
客户每天有海量的餐饮物资需要入库，但是手动的填写多维表格又非常麻烦

 (语音识别+图片OCR) x OpenAI的信息摘要能力 = 多模态的数据录入


 ```mermaid
stateDiagram-v2
    BOT --> 语音录入
    语音录入 --> 飞书语音转文字
    飞书语音转文字 --> QA_FewShot
    QA_FewShot --> OpenAI_Summary
    OpenAI_Summary --> 操作指令_JSON
    操作指令_JSON --> 读写多维表格

    BOT --> 票据图片
    票据图片 --> 阿里云票据OCR
    阿里云票据OCR --> QA_FewShot

    读取多维表格 --> 获取已有数据库信息 
    获取已有数据库信息 --> QA_FewShot

```

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

## 复现部署
1. 导入MVP环境的[多维表格至本地](https://fork-way.feishu.cn/base/Norqbd3anazOMlsceoEcbeIMn2g?table=tblghsq6TY3XktYG&view=vewr4Kwn8O)
2. 复制一份`config.yaml`文件，填写对应的配置
    ```json5
	#飞书BOT配置
	APP_ID: cli_a4172ca48e7b9013
	APP_SECRET: tWFwzQNzkD..
	APP_ENCRYPT_KEY: 39XCQUA00uqeDC..
	APP_VERIFICATION_TOKEN: CYyoRCIz..
	BOT_NAME: ChatGPT
	
	# openai 配置
	OPENAI_MODEL: gpt-3.5-turbo
	# openAI 最大token数 默认为2000
	OPENAI_MAX_TOKENS: 2000
	OPENAI_KEY: sk-zTeWCJSr...
	
	# mvp多维表格 
	BITABLE_HOST: fork-way.feishu.cn
	BITABLE_APP_TOKEN: Norqbd3anazOMlsceoEcbeIMn2g
	#产品数据库tableID
	PRODUCT_BITABLE_ID: tblVY2JwWJfbzWzS
	#供应商tableID
	SUPPLIER_BITABlE_ID: tbltrRO7DJGXd1ZJ
	#目标写入tableID及其视图ID
	RECEIPT_BITABLE_ID: tblghsq6TY3XktYG
	RECEIPT_BITABLE_VIEW: vewr4Kwn8O
	
	# 阿里云 OCR https://ai.aliyun.com/ocr/invoice
	ALI_ACCESSKEY_ID: LTAI5tEBvzfgfjZSy...
	ALI_ACCESSKEY_SECRET: DxpKnI64q31630....
	
	# 服务器配置
	HTTP_PORT: 9001
	HTTPS_PORT: 9002
	
	```

3. 此应用整体基于飞书-OpenAI项目改造完成，初[原项目配置](https://github.com/ConnectAI-E/Feishu-OpenAI#%E8%AF%A6%E7%BB%86%E9%85%8D%E7%BD%AE%E6%AD%A5%E9%AA%A4)之外，还需下列额外操作
	```
	事件权限-查看、评论、编辑和管理多维表格 
	进入多维表格-更多-添加文档应用-添加此应用
	```
<img width="420" alt="image" src="https://github.com/ConnectAI-E/Base-EasyWrite/assets/50035229/d2c0511d-0df2-4e74-a3de-7e7a889f9d96">

4. 输入测试文本：`收到了大白菜 30斤 一共100块`

## 下一步思考

如何让数据录入的场景更加通用，从而减少定制化的需求呢？

下一步一定是OA-Agent~

以录入助手为例，其本身分为两个agent：
```
1.数据收集agent，作用：产生约定的json字段，他下辖
- 语音收集agent -作用：自然语音-->约定的json字段
- 票据收集agent-作用：票据图片-->约定的json字段
2.多维表格录入agent，作用：自然语言或者约定的指令-->读写多维表格
```

那我们在定制什么呢? 我们其实就在定制QA

所有的Agent都接受QA作为历史学习经验的积累，才可以上岗

Agent的学习方式现在看来至少包括：
- 提示词 Prompt、
- few-short的QA、
- context-window向量索引上下文的QA
- 利用Lora微调模型本身


