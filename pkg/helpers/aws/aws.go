/*
Copyright © 2020 Kenneth Pianeta <kjpianeta@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package aws

import (
	"fmt"
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)
const (
	INSTANCE_STATE_PENDING       int64 = 0
	INSTANCE_STATE_RUNNING       int64 = 16
	INSTANCE_STATE_SHUTTING_DOWN int64 = 32
	INSTANCE_STATE_TERMINATED    int64 = 48
	INSTANCE_STATE_STOPPING      int64 = 64
	INSTANCE_STATE_STOPPED       int64 = 80
)
func GetInstanceStateList(deploymentTag string, currentState string, ec2client *ec2.EC2) ([]string, error) {

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	ec2Svc := ec2.New(sess)
	params := &ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("tag:Deployment"),
				Values: []*string{aws.String(deploymentTag)},
			},
			{
				Name:   aws.String("instance-state-name"),
				Values: []*string{aws.String(currentState)},
			},
		},
	}

	instanceResp, err := ec2Svc.DescribeInstances(params)
	instanceIds := make([]string, 0, 10)
	if err != nil {
		fmt.Println("Error", err)
	} else {
		if len(instanceResp.Reservations) > 0 {
			reservation := instanceResp.Reservations[0]
			if len(reservation.Instances) > 0 {
				for _, reservation := range instanceResp.Reservations {
					for _, instance := range reservation.Instances {
						instanceIds = append(instanceIds, *instance.InstanceId)
					}
				}
			}
		}
	}
	return instanceIds, err
}

func GetEc2Client() *ec2.EC2 {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	ec2Svc := ec2.New(sess)
	return ec2Svc
}
func StartInstance(instanceids []string, ec2client *ec2.EC2) error {
	// Create our struct to hold everything
	instanceReq := ec2.StartInstancesInput{
		InstanceIds: []*string{},
	}

	// Loop through our input array and add them to our struct, converting them to the string pointer required by the SDK
	for _, id := range instanceids {
		instanceReq.InstanceIds = append(instanceReq.InstanceIds, aws.String(id))
	}

	//Make the request to start all the instances, returning an error if we got one.
	instanceResp, err := ec2client.StartInstances(&instanceReq)
	if err != nil {
		return err
	}

	// The number of instances we got back should be the same as how many we requested.
	if len(instanceResp.StartingInstances) != len(instanceids) {
		return errors.New("the total number of started instances did not match the request")
	}

	// Finally, let's loop through all of the responses and see they started.
	// We'll store each ID in a string so we can build a good error and use it to see later if we had any not started
	allStarted := ""

	// Loop through all the instances and check the state
	for _, instance := range instanceResp.StartingInstances {
		//if *instance.CurrentState.Code != INSTANCE_STATE_RUNNING && *instance.CurrentState.Code != INSTANCE_STATE_PENDING {
		if *instance.CurrentState.Code != INSTANCE_STATE_RUNNING {
			allStarted = allStarted + " " + *instance.InstanceId
		}
	}

	// If we found some that didn't start then return the error
	if allStarted != "" {
		return errors.New("The following instances did not start: " + allStarted)
	}

	// Else let's return nil for success
	return nil
}

func StopInstance(instanceids []string, ec2client *ec2.EC2) error {
	// Create our struct to hold everything
	instanceReq := ec2.StopInstancesInput{
		InstanceIds: []*string{},
	}

	// Loop through our input array and add them to our struct, converting them to the string pointer required by the SDK
	for _, id := range instanceids {
		instanceReq.InstanceIds = append(instanceReq.InstanceIds, aws.String(id))
	}

	//Make the request to stop all the instances, returning an error if we got one.
	instanceResp, err := ec2client.StopInstances(&instanceReq)
	if err != nil {
		return err
	}

	// The number of instances we got back should be the same as how many we requested.
	if len(instanceResp.StoppingInstances) != len(instanceids) {
		return errors.New("the total number of stopped instances did not match the request")
	}

	// Finally, let's loop through all of the responses and see they started.
	// We'll store each ID in a string so we can build a good error and use it to see later if we had any not stopped
	allStopped := ""

	// Loop through all the instances and check the state
	for _, instance := range instanceResp.StoppingInstances {
		if *instance.CurrentState.Code != INSTANCE_STATE_STOPPED && *instance.CurrentState.Code != INSTANCE_STATE_STOPPING{
			allStopped = allStopped + " " + *instance.InstanceId
		}
	}

	// If we found some that didn't stop then return the error
	if allStopped != "" {
		return errors.New("The following instances did not stop: " + allStopped)
	}

	// Else let's return nil for success
	return nil
}
