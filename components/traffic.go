package components

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	lighthouse "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/lighthouse/v20200324"
)

var credential = common.NewCredential(
	Config.TencentKeyOne,
	Config.TencentKeyTwo,
)

type TrafficPackage struct {
	TrafficPackageId        string
	TrafficUsed             int
	TrafficPackageTotal     int
	TrafficPackageRemaining int
	TrafficOverflow         int
	Starttime               string
	Endtime                 string
	DeadLine                string
	Status                  string
}

type InstanceTrafficPackage struct {
	InstanceId        string
	TrafficPackageSet []*TrafficPackage
}

type TencentLightHouseTraffic struct {
	TotalCount                int
	RequestId                 string
	InstanceTrafficPackageSet []*InstanceTrafficPackage
}

type TencentLighthouseTrafficResponse struct {
	Response *TencentLightHouseTraffic `json:"Response"`
}

func TencentLighthouseTrafficUsage() (*TencentLighthouseTrafficResponse, error) {
	// message := ""
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "lighthouse.tencentcloudapi.com"
	client, _ := lighthouse.NewClient(credential, "ap-hongkong", cpf)
	request := lighthouse.NewDescribeInstancesTrafficPackagesRequest()
	response, err := client.DescribeInstancesTrafficPackages(request)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		fmt.Printf("An API error has returned: %s", err)
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	resString := response.ToJsonString()

	result := new(TencentLighthouseTrafficResponse)
	err = json.Unmarshal([]byte(resString), result)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	// log.Println(result.Response.InstanceTrafficPackageSet[0].TrafficPackageSet[0].TrafficUsed)
	return result, nil
}

func TencentLighthouseTrafficUsageShow() string {
	t, err := TencentLighthouseTrafficUsage()
	if err != nil {
		return err.Error()
	}
	instanceNum := t.Response.TotalCount
	instanceID := t.Response.InstanceTrafficPackageSet[0].InstanceId
	trafficPackageID := t.Response.InstanceTrafficPackageSet[0].TrafficPackageSet[0].TrafficPackageId
	trafficUsed := float64(t.Response.InstanceTrafficPackageSet[0].TrafficPackageSet[0].TrafficUsed) / (1024 * 1024 * 1024)
	trafficAll := float64(t.Response.InstanceTrafficPackageSet[0].TrafficPackageSet[0].TrafficPackageTotal) / (1024 * 1024 * 1024)
	msg := fmt.Sprintf("This account has instance: %d\nInstance ID:%s\nTrafficPackageID:%s\nTrafficUsed:%.3fGiB\nTrafficUsedPercent:%.3f%%", instanceNum, instanceID, trafficPackageID, trafficUsed, 100*trafficUsed/trafficAll)
	log.Println(msg)

	return msg
}
