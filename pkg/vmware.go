package pkg

import (
	"context"
	"fmt"
	"net/url"

	log "github.com/sirupsen/logrus"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/performance"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

// VMware scracpe object
type VMware struct {
	vsphereHost     string
	vsphereUser     string
	vspherePassword string
	insecure        bool
	client          govmomi.Client
	ctx             context.Context
	perfMgr         *performance.Manager
}

// Instance is vmware vm local object
type Instance struct {
	uuid string
}

// NewVMware return new VMware struct
func NewVMware(vsphereHost, vsphereUser, vspherePassword string, insecure bool) (*VMware, error) {

	v := VMware{
		vsphereHost:     vsphereHost,
		vsphereUser:     vsphereUser,
		vspherePassword: vspherePassword,
		insecure:        insecure,
		ctx:             context.Background(),
	}

	u, err := url.Parse(fmt.Sprintf("https://%s:%s@%s/sdk", v.vsphereUser, v.vspherePassword, v.vsphereHost))
	if err != nil {
		log.Errorf("parse vmware connection failed (%s).", err)
		return nil, err
	}
	client, err := govmomi.NewClient(v.ctx, u, v.insecure)
	if err != nil {
		log.Errorf("init vmware connection failed (%s)", err)
		return nil, err
	}
	v.client = *client
	v.perfMgr = performance.NewManager(client.Client)
	return &v, nil
}

// GetDataCenterList returns datacenter object list
func (v *VMware) GetDataCenterList() (*find.Finder, []*object.Datacenter, error) {

	finder := find.NewFinder(v.client.Client, true)
	dataceneters, err := finder.DatacenterList(v.ctx, "*")
	return finder, dataceneters, err
}

// GetVirtaulMachineByDataCenter returns all virtual machine by dc
func (v *VMware) GetVirtaulMachineByDataCenter(finder find.Finder, dataceneter *object.Datacenter) (
	[]*object.VirtualMachine, error) {

	finder.SetDatacenter(dataceneter)
	vms, err := finder.VirtualMachineList(v.ctx, "*")
	return vms, err
}

// GetVirtaulMachineBasicsInfo returns virtalmachine basics infomartion
func (v *VMware) GetVirtaulMachineBasicsInfo(vms []*object.VirtualMachine) ([]mo.VirtualMachine, error) {

	objs := []types.ManagedObjectReference{}
	for _, vm := range vms {
		objs = append(objs, vm.Reference())
	}
	virtualMachineInfoList := []mo.VirtualMachine{}

	pc := property.DefaultCollector(v.client.Client)
	err := pc.Retrieve(v.ctx, objs, vmProperties, &virtualMachineInfoList)

	return virtualMachineInfoList, err
}

// GenLocalInstance returns instance
func (v *VMware) GenLocalInstance(vm mo.VirtualMachine, dcName string) *Instance {
	summary := vm.Summary

	instance := Instance{
		uuid: summary.Config.InstanceUuid,
	}
	return &instance
}

// GetVirtaulMachinePerformance returns virtal machine performance
func (v *VMware) GetVirtaulMachinePerformance(instanceUUID string,
	objs []types.ManagedObjectReference, spec types.PerfQuerySpec,
	metrics []string) ([]performance.EntityMetric, error) {

	series, err := v.perfMgr.SampleByName(v.ctx, spec, metrics, objs)
	if err != nil {
		log.Errorf("get virtal machine(%s)'s series failed (%s).", instanceUUID, err)
		return nil, err
	}

	metircSeries, err := v.perfMgr.ToMetricSeries(v.ctx, series)
	if err != nil {
		log.Errorf("vm(%s) to metric series failed (%s).", instanceUUID, err)
		return nil, err
	}

	return metircSeries, nil
}

// CounterInfo returns counter info list
func (v *VMware) CounterInfo() ([]types.PerfCounterInfo, error) {
	counters, err := v.perfMgr.CounterInfo(v.ctx)
	if err != nil {
		return counters, err
	}
	return counters, err
}

// DeclareSpec returns spec
func (v *VMware) DeclareSpec() types.PerfQuerySpec {

	spec := types.PerfQuerySpec{
		Format:     string(types.PerfFormatNormal),
		MaxSample:  1,
		IntervalId: 20,
	}
	return spec
}

// DeclareMetrics returns vm performance metrics
func (v *VMware) DeclareMetrics() []string {

	metrics := []string{}
	for key := range vmMetricMap {
		metrics = append(metrics, key)
	}
	return metrics
}

// destroy vmware client connection
func (v *VMware) destroy() {

	err := v.client.Logout(v.ctx)
	if err != nil {
		log.Errorf("destroy vmware client failed (%s).", err)
	}
}

// LableValues returns metrics label values
func (ins *Instance) LableValues(extend string) []string {
	values := []string{
		ins.uuid,
		extend,
	}
	return values
}
