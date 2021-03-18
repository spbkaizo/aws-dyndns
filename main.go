package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
)

var target = "kaizo.org" // A record to add to external IP from called host
var zoneid = "Z04587871QTB3RYSR6JZM"

const wtfurl = "https://wtfismyip.com/json"

// IPResponse
// Returns from wtfismyip.com/json
type IPResponse struct {
	YourFuckingCountryCode string `json:"YourFuckingCountryCode"`
	YourFuckingHostname    net.IP `json:"YourFuckingHostname"`
	YourFuckingIPAddress   net.IP `json:"YourFuckingIPAddress"`
	YourFuckingISP         string `json:"YourFuckingISP"`
	YourFuckingLocation    string `json:"YourFuckingLocation"`
	YourFuckingTorExit     bool   `json:"YourFuckingTorExit"`
}

func main() {

	var wtfresp IPResponse
	resp, err := http.Get(wtfurl)
	if err != nil {
		log.Fatalf("FATAL: %v", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("FATAL: %v", err)
	}

	if err := json.Unmarshal(body, &wtfresp); err != nil {
		panic(err)
	}

	myname, err := os.Hostname()
	if err != nil {
		myname = "amnesiac"
	}

	svc := route53.New(session.New())
	input := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &route53.ChangeBatch{
			Changes: []*route53.Change{
				{
					Action: aws.String("UPSERT"),
					ResourceRecordSet: &route53.ResourceRecordSet{
						Name: aws.String(target),
						ResourceRecords: []*route53.ResourceRecord{
							{
								Value: aws.String(wtfresp.YourFuckingIPAddress.String()),
							},
						},
						TTL:  aws.Int64(60),
						Type: aws.String("A"),
					},
				},
			},
			Comment: aws.String("Record updated by " + myname),
		},
		HostedZoneId: aws.String(zoneid),
	}

	result, err := svc.ChangeResourceRecordSets(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case route53.ErrCodeNoSuchHostedZone:
				fmt.Println(route53.ErrCodeNoSuchHostedZone, aerr.Error())
			case route53.ErrCodeNoSuchHealthCheck:
				fmt.Println(route53.ErrCodeNoSuchHealthCheck, aerr.Error())
			case route53.ErrCodeInvalidChangeBatch:
				fmt.Println(route53.ErrCodeInvalidChangeBatch, aerr.Error())
			case route53.ErrCodeInvalidInput:
				fmt.Println(route53.ErrCodeInvalidInput, aerr.Error())
			case route53.ErrCodePriorRequestNotComplete:
				fmt.Println(route53.ErrCodePriorRequestNotComplete, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		return
	}

	fmt.Println(result)
}
