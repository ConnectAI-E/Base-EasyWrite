package ocr

import (
	"encoding/json"
	"inventory-manager/initialization"
	"os"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	ocr_api20210707 "github.com/alibabacloud-go/ocr-api-20210707/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
)

var ocrClient *ocr_api20210707.Client

func GetOcrClient(accessKeyId, accessKeySecret string) (*ocr_api20210707.Client, error) {
	if ocrClient == nil {
		config := &openapi.Config{
			AccessKeyId:     tea.String(accessKeyId),
			AccessKeySecret: tea.String(accessKeySecret),
		}
		config.Endpoint = tea.String("ocr-api.cn-hangzhou.aliyuncs.com")
		client, err := ocr_api20210707.NewClient(config)
		if err != nil {
			return nil, err
		}
		ocrClient = client
	}
	return ocrClient, nil
}

func RecognizeTableOcr(config initialization.Config, picPath string) (string, error) {
	client, _err := GetOcrClient(config.AliAccessKeyId, config.AliAccessKeySecret)
	if _err != nil {
		return "", _err
	}

	f, err := os.Open(picPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	recognizeTableOcrRequest := &ocr_api20210707.RecognizeTableOcrRequest{
		Body: f,
		// Url: tea.String("https://www.fsxlsoft.com/upload/201811/1542940677331387.png"),
	}
	runtime := &util.RuntimeOptions{}
	res, _err := client.RecognizeTableOcrWithOptions(recognizeTableOcrRequest, runtime)
	if _err != nil {
		return "", _err
	}

	if res == nil || res.Body == nil || res.Body.Data == nil {
		return "", nil
	}

	var data *RespData
	_err = json.Unmarshal([]byte(*res.Body.Data), &data)
	if _err != nil {
		return "", _err
	}

	return data.Content, nil
}

type RespData struct {
	Content string `json:"content"`
}
