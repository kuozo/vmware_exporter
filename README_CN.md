# VMWware Exporter

> [中文版本](https://github.com/klnchu/vmware_exporter/blob/master/README_CN.md) | [English](https://github.com/klnchu/vmware_exporter/blob/master/README.md)

通过 VMware SDK 获取 VMware 符合 Prometheus 监控性能数据

## 准备工作

> VMware 设置项中设置 ```vpxd.stats.maxquerymetrics``` 的值为 -1，解除调用接口次数限制

## Golang

* 版本： 1.11 版本以及以上

## 说明

* 原始的百分比值为已经乘以 100 以后的值 


## 参考链接

* [VMware vSphere Performance Metrics](https://pubs.vmware.com/vsphere-4-esx-vcenter/index.jsp?topic=/com.vmware.vsphere.bsa.doc_40/vc_admin_guide/performance_metrics/c_performance_metrics.html)
* [Vwarem Managed Object - PerformanceManager](https://www.vmware.com/support/developer/converter-sdk/conv61_apireference/vim.PerformanceManager.html)



