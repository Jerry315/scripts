package handle

import (
	"encoding/json"
	"fmt"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/alidns"
	"log"
)

func ParseResponse(result []byte) {
	// 解析返回的数据，输出友好的json格式
	var f interface{}
	err := json.Unmarshal(result, &f)
	if err != nil {
		fmt.Println(err)
	}
	m := f.(map[string]interface{})
	b, err := json.MarshalIndent(m, "", "\t")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(b))
}

func GetDomainRecords(client *alidns.Client, DomainName string, PageNumber, PageSize requests.Integer) {
	/*
	根据传入参数获取指定主域名的所有解析记录列表。
	查询可以指定域名（DomainName）、页码（PageNumber）和每页的数量（PageSize）来获取域名的解析列表。
	查询可以指定解析记录的主机记录关键字（RRKeyWord）、解析类型关键字（TypeKeyWord）或者记录值的关键字（ValueKeyWord）
	来查询含有该关键字的解析列表。
	解析列表的默认排序方式是按照解析添加的时间从新到旧排序的。
	*/
	request := alidns.CreateDescribeDomainRecordsRequest()
	request.DomainName = DomainName
	request.PageNumber = PageNumber
	request.PageSize = PageSize
	response, err := client.DescribeDomainRecords(request)
	if err != nil {
		log.Printf("GetDomainRecords failed: #%v",err)
	}
	if err != nil {
		fmt.Print(err.Error())
	}
	result := response.BaseResponse.GetHttpContentBytes()
	ParseResponse(result)
}

func GetDomainRecordInfo(client *alidns.Client, RecordId string) {
	/*
	根据RecordId获取记录的详细信息
	*/
	request := alidns.CreateDescribeDomainRecordInfoRequest()
	request.RecordId = RecordId
	response, err := client.DescribeDomainRecordInfo(request)
	if err != nil {
		log.Printf("GetDomainRecordInfo failed: #%v",err)
	}
	if err != nil {
		fmt.Print(err.Error())
	}
	result := response.BaseResponse.GetHttpContentBytes()
	ParseResponse(result)
}

func GetSubDomainRecord(client *alidns.Client, SubDomain string, PageNumber, PageSize requests.Integer) {
	/*
	根据传入参数获取某个固定子域名的所有解析记录列表
	*/
	request := alidns.CreateDescribeSubDomainRecordsRequest()
	request.SubDomain = SubDomain
	request.PageSize = PageSize
	request.PageNumber = PageNumber
	response, err := client.DescribeSubDomainRecords(request)
	if err != nil {
		log.Printf("GetSubDomainRecord failed: #%v",err)
	}
	result := response.BaseResponse.GetHttpContentBytes()
	ParseResponse(result)
}

func AddDomainRecord(client *alidns.Client, DomainName, RR, Value, _Type string) {
	/*
	根据传入参数添加解析记录
	*/
	request := alidns.CreateAddDomainRecordRequest()
	request.DomainName = DomainName
	request.RR = RR
	request.Value = Value
	request.Type = _Type
	response, err := client.AddDomainRecord(request)
	if err != nil {
		log.Printf("AddDomainRecord failed: #%v",err)
	}
	result := response.BaseResponse.GetHttpContentBytes()
	ParseResponse(result)
}

func DelDomainRecord(client *alidns.Client, RecordId string) {
	/*
	根据传入参数删除解析记录
	*/
	request := alidns.CreateDeleteDomainRecordRequest()
	request.RecordId = RecordId
	response, err := client.DeleteDomainRecord(request)
	if err != nil {
		log.Printf("DeldomainRecord failed: #%v",err)
	}
	result := response.BaseResponse.GetHttpContentBytes()
	ParseResponse(result)
}

func UpdateDomainRecord(client *alidns.Client, RecordId, RR, _Type, Value string) {
	/*
	根据传入参数修改解析记录
	*/
	request := alidns.CreateUpdateDomainRecordRequest()
	request.RecordId = RecordId
	request.RR = RR
	request.Type = _Type
	request.Value = Value
	response, err := client.UpdateDomainRecord(request)
	if err != nil {
		log.Printf("UpdateDomainRecord failed: #%v",err)
	}
	result := response.BaseResponse.GetHttpContentBytes()
	ParseResponse(result)
}

func SetDomainRecord(client *alidns.Client,RecordId,Status string)  {
	/*
	根据传入参数获取设置解析记录状态
	*/
	request := alidns.CreateSetDomainRecordStatusRequest()
	request.RecordId = RecordId
	request.Status = Status
	response,err := client.SetDomainRecordStatus(request)
	if err != nil {
		log.Printf("SetDomainRecord failed: #%v",err)
	}
	result := response.BaseResponse.GetHttpContentBytes()
	ParseResponse(result)
}