package client

import (
	"errors"
	"github.com/ReSTARTR/ec2-ls-hosts/creds"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"regexp"
	"sort"
	"strings"
)

type FormattedEc2 struct {
	Values []string
	Index  int
}

type Ec2s []*FormattedEc2

func (e Ec2s) Len() int {
	return len(e)
}

func (e Ec2s) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}

func (e Ec2s) Less(i, j int) bool {
	return e[i].Values[e[i].Index] < e[j].Values[e[j].Index]
}

type Writer interface {
	SetHeader(s []string)
	Append(s []string)
	Render()
}

var (
	defaultFields = []string{
		"tag:Name",
		"instance-id",
		"private-ip",
		"public-ip",
		"instance-state",
	}
)

type Options struct {
	Filters     map[string]string
	TagFilters  map[string]string
	Fields      []string
	Region      string
	Credentials string
	Profiles    string
	SortField   int
	Noheader    bool
}

func NewOptions() *Options {
	opt := &Options{}
	opt.Filters = make(map[string]string)
	opt.TagFilters = make(map[string]string)
	opt.SortField = -1
	return opt
}

func (o *Options) FieldNames() []string {
	if len(o.Fields) > 1 {
		return o.Fields
	}
	return defaultFields
}

func Describe(o *Options, w Writer) error {
	// build queries
	config := &aws.Config{Region: aws.String(o.Region)}
	credentials_list, err := creds.SelectCredentials(o.Credentials, o.Profiles)
	if err != nil {
		return err
	}
	instances := make([]*ec2.Instance,0)
	for _, credentials := range credentials_list {
		config.Credentials = credentials
		list, err := listInstances(config, o.Filters, o.TagFilters)
		if err != nil {
			return err
		}
		instances = append(instances, list...)
	}

	if o.Noheader == false {
		w.SetHeader(o.FieldNames())
	}
	formatedInstances := make([]*FormattedEc2, 0)
	for _, inst := range instances {
		values := formatInstance(inst, o.FieldNames())
		fec2 := &FormattedEc2{
			Values: values,
			Index: o.SortField,
		}
		formatedInstances = append(formatedInstances, fec2)
	}
	if o.SortField != -1 {
		sort.Sort(Ec2s(formatedInstances))
	}
	for _, inst := range formatedInstances {
		w.Append(inst.Values)
	}
	w.Render()

	return nil
}

func listInstances(config *aws.Config, filters map[string]string, tagFilters map[string]string) ([]*ec2.Instance, error) {
	svc := ec2.New(session.New(), config)

	// call aws api
	options := &ec2.DescribeInstancesInput{}
	for k, v := range filters {
		options.Filters = append(options.Filters, &ec2.Filter{
			Name:   aws.String(k),
			Values: []*string{aws.String(v)},
		})
	}
	for k, v := range tagFilters {
		options.Filters = append(options.Filters, &ec2.Filter{
			Name:   aws.String("tag:" + k),
			Values: []*string{aws.String(v)},
		})
	}

	// show info
	resp, err := svc.DescribeInstances(options)
	if err != nil {
		return nil, err
	}
	if len(resp.Reservations) == 0 {
		return nil, errors.New("Not Found")
	}
	instances := make([]*ec2.Instance, 0)
	for idx, _ := range resp.Reservations {
		instances = append(instances, resp.Reservations[idx].Instances...)
	}
	return instances, nil
}

func formatInstance(inst *ec2.Instance, fields []string) []string {
	// fetch IPs
	var privateIps []string
	var publicIps []string
	for _, nic := range inst.NetworkInterfaces {
		for _, privateIp := range nic.PrivateIpAddresses {
			privateIps = append(privateIps, *privateIp.PrivateIpAddress)
			if privateIp.Association != nil {
				publicIps = append(publicIps, *privateIp.Association.PublicIp)
				break
			}
		}
	}

	// fetch tags
	// NOTE: *DO NOT* support multiple tag values
	tags := make(map[string]string, 5)
	for _, tag := range inst.Tags {
		tags[*tag.Key] = *tag.Value
	}

	var values []string
	for _, c := range fields {
		switch c {
		case "instance-id":
			values = append(values, *inst.InstanceId)
		case "private-ip":
			values = append(values, strings.Join(privateIps, ","))
		case "public-ip":
			values = append(values, strings.Join(publicIps, ","))
		case "launch-time":
			values = append(values, inst.LaunchTime.String())
		case "instance-state":
			values = append(values, *inst.State.Name)
		default:
			// extract key-values as tag string
			matched, err := regexp.Match("tag:.+", []byte(c))
			if err == nil && matched {
				kv := strings.Split(c, ":")
				key := strings.Join(kv[1:len(kv)], ":")
				if v, ok := tags[key]; ok {
					values = append(values, v)
				}
			}
		}
	}

	return values
}
