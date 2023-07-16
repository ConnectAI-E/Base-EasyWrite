package lark

import (
	"context"
	"encoding/json"
	"fmt"
	"inventory-manager/initialization"
	"strconv"
	"strings"
	"time"

	larkbitable "github.com/larksuite/oapi-sdk-go/v3/service/bitable/v1"
)

func GetAllProducts(config initialization.Config) (map[string][2]string, error) {
	client := initialization.GetLarkClient()
	req := larkbitable.NewListAppTableRecordReqBuilder().
		AppToken(config.BitableAppToken).
		TableId(config.ProductBitableId).
		FieldNames(`["产品名称","主供应商"]`).
		TextFieldAsArray(true).
		UserIdType("user_id").
		Build()
	resp, err := client.Bitable.AppTableRecord.List(context.Background(), req)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	if !resp.Success() {
		fmt.Println(resp.Code, resp.Msg, resp.RequestId())
		return nil, err
	}

	var r Response
	err = json.Unmarshal(resp.ApiResp.RawBody, &r)
	if err != nil {
		fmt.Println("Failed to deserialize string:", err)
		return nil, err
	}

	result := make(map[string][2]string)
	for _, item := range r.Data.Items {
		companyName := ""
		if key, ok := item.Fields["主供应商"]; ok {
			if len(key) > 0 {
				companyName = key[0].Text
			}
		}
		if key, ok := item.Fields["产品名称"]; ok {
			if len(key) > 0 {
				result[key[0].Text] = [2]string{item.RecordId, companyName}
			}
		}
	}
	return result, nil
}

func GetAllSuppliers(config initialization.Config) (map[string]string, error) {
	client := initialization.GetLarkClient()
	req := larkbitable.NewListAppTableRecordReqBuilder().
		AppToken(config.BitableAppToken).
		TableId(config.SupplierBitableId).
		FieldNames(`["供应商名称"]`).
		TextFieldAsArray(true).
		UserIdType("user_id").
		Build()
	resp, err := client.Bitable.AppTableRecord.List(context.Background(), req)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	if !resp.Success() {
		fmt.Println(resp.Code, resp.Msg, resp.RequestId())
		return nil, err
	}

	var r Response
	err = json.Unmarshal(resp.ApiResp.RawBody, &r)
	if err != nil {
		fmt.Println("Failed to deserialize string:", err)
		return nil, err
	}

	result := make(map[string]string, 0)
	for _, item := range r.Data.Items {
		if key, ok := item.Fields["供应商名称"]; ok {
			if len(key) > 0 {
				result[key[0].Text] = item.RecordId
			}
		}
	}
	return result, nil
}

type ReceiptRecord struct {
	Date        string  `json:"date"`
	ProductName string  `json:"productName"`
	CompanyName string  `json:"companyName"`
	TotalNumber float64 `json:"totalNumber"`
	TotalPrice  float64 `json:"totalPrice"`
	Remark      string  `json:"remark"`
}

func CreateTableRecord(config initialization.Config, recordList []*ReceiptRecord, products map[string][2]string, suppliers map[string]string) (string, string, error) {
	replayItems := make([]map[string]interface{}, 0)
	records := make([]*larkbitable.AppTableRecord, 0)

	for _, item := range recordList {
		t, err := time.Parse("2006/01/02", item.Date)
		if err != nil {
			t = time.Now()
			fmt.Println("time.Parse err:", err)
		}
		timestamp := t.UnixMilli()

		var rid string
		defaultCompanyName := ""
		ridAndCompanyName, ok := products[item.ProductName]
		if !ok {
			arr := getProductByName(products, item.ProductName)
			if len(arr) == 3 {
				item.ProductName = arr[0]
				rid = arr[1]
				defaultCompanyName = arr[2]
			}
		} else {
			rid = ridAndCompanyName[0]
			defaultCompanyName = ridAndCompanyName[1]
		}

		if len(rid) > 0 {
			isConsistent, companyName, companyRid := getCompany(suppliers, defaultCompanyName, item.CompanyName)
			record := map[string]interface{}{
				"日期":   timestamp,
				"产品名称": []string{rid},
				"实收":   item.TotalNumber,
				"金额":   item.TotalPrice,
			}
			hint := ""
			if !isConsistent {
				if len(companyRid) > 0 {
					record["非主供应商"] = true
					record["本次供应商"] = []string{companyRid}
				} else {
					hint = "（无效字段，仍使用主供应商）"
				}
			}
			records = append(records, larkbitable.NewAppTableRecordBuilder().Fields(record).Build())
			replayItem := map[string]interface{}{
				// "链接":   rid,
				"日期":   item.Date,
				"产品名称": item.ProductName,
				"实收":   item.TotalNumber,
				"金额":   item.TotalPrice,
			}
			if !isConsistent {
				replayItem["供应商"] = companyName + hint
			}
			replayItems = append(replayItems, replayItem)
		} else {
			records = append(records, larkbitable.NewAppTableRecordBuilder().Fields(map[string]interface{}{
				"日期":   timestamp,
				"产品名称": []string{"recnqXqgdq"},
				"实收":   item.TotalNumber,
				"金额":   item.TotalPrice,
				"备注":   item.ProductName,
			}).Build())
			replayItems = append(replayItems,
				map[string]interface{}{
					// "链接":   "recnqXqgdq",
					"日期":   item.Date,
					"产品名称": "员工餐",
					"实收":   item.TotalNumber,
					"金额":   item.TotalPrice,
					"备注":   item.ProductName,
				})
		}
	}

	client := initialization.GetLarkClient()
	req := larkbitable.NewBatchCreateAppTableRecordReqBuilder().
		AppToken(config.BitableAppToken).
		TableId(config.ReceiptBitableId).
		Body(larkbitable.NewBatchCreateAppTableRecordReqBodyBuilder().
			Records(records).
			Build()).
		Build()
	resp, err := client.Bitable.AppTableRecord.BatchCreate(context.Background(), req)
	if err != nil {
		fmt.Println(err)
		return "", "", err
	}
	if !resp.Success() {
		return "", "", fmt.Errorf(resp.Msg)
	}

	result := make([]string, 0)
	for _, item := range replayItems {
		bytes, _ := json.Marshal(item)
		result = append(result, strings.ReplaceAll(string(bytes), `"`, ""))
	}

	firstLink := ""
	if len(resp.Data.Records) > 0 {
		rid := *resp.Data.Records[0].RecordId
		firstLink = `https://` + config.BitableHost + `/base/` + config.BitableAppToken + `?table=` + config.ReceiptBitableId + `&view=` + config.ReceiptBitableView + `&record=` + rid
	}
	// for _, record := range resp.Data.Records {
	// 	productRid := record.Fields["产品名称"].([]interface{})[0].(string)
	// 	for _, item := range replayItems {
	// 		if item["链接"] == productRid && item["实收"] == record.Fields["实收"] && item["金额"] == record.Fields["金额"] {
	// 			rid := *record.RecordId
	// 			item["链接"] = `https://` + config.BitableHost + `/base/` + config.BitableAppToken + `?table=` + config.ReceiptBitableId + `&view=` + config.ReceiptBitableView +`&field=fld9H5IdjI&record=` + rid
	// 			bytes, _ := json.Marshal(item)
	// 			result = append(result, strings.ReplaceAll(string(bytes), `"`, ""))
	// 		}
	// 	}
	// }
	resultStr := ""
	for i, item := range result {
		resultStr += "**" + strconv.Itoa(i+1) + ".** " + item + "\n"
	}

	return resultStr, firstLink, nil
}

func getProductByName(products map[string][2]string, name string) []string {
	for product, ridAndCompanyName := range products {
		if strings.Contains(product, name) {
			return []string{product, ridAndCompanyName[0], ridAndCompanyName[1]}
		}
	}
	return []string{}
}

func getCompany(supplier map[string]string, defaultCompany, recognizedName string) (bool, string, string) {
	if len(recognizedName) == 0 || strings.Contains(defaultCompany, recognizedName) {
		return true, defaultCompany, ""
	}
	if rid, ok := supplier[recognizedName]; ok {
		return false, recognizedName, rid
	}
	for name, rid := range supplier {
		if strings.Contains(name, recognizedName) {
			return false, name, rid
		}
	}
	return false, recognizedName, ""
}

type Response struct {
	Code int         `json:"code"`
	Data DataWrapper `json:"data"`
}

type DataWrapper struct {
	Total int64      `json:"total"`
	Items []*Product `json:"items"`
}
type Product struct {
	RecordId string                   `json:"record_id"`
	Fields   map[string][]ProductName `json:"fields"`
}

type ProductName struct {
	Text string `json:"text"`
	Type string `json:"type"`
}
