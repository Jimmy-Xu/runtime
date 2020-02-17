// Copyright (c) 2017-2018 Intel Corporation
// Copyright (c) 2018 Huawei Corporation
//
// SPDX-License-Identifier: Apache-2.0
//

package manager

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/kata-containers/runtime/virtcontainers/device/config"
)

const (
	vfioPath = "/dev/vfio/"

	PCIClassCodeDisplayController = "0x03"
)

// isVFIO checks if the device provided is a vfio group.
func isVFIO(hostPath string) bool {
	// Ignore /dev/vfio/vfio character device
	if strings.HasPrefix(hostPath, filepath.Join(vfioPath, "vfio")) {
		return false
	}

	if strings.HasPrefix(hostPath, vfioPath) && len(hostPath) > len(vfioPath) {
		return true
	}

	return false
}

// isBlock checks if the device is a block device.
func isBlock(devInfo config.DeviceInfo) bool {
	return devInfo.DevType == "b"
}

func IsVFIODisplay(hostPath string) (bool, error) {
	if !isVFIO(hostPath) {
		return false, nil
	}
	iommuDevicesPath := filepath.Join(config.SysIOMMUPath, filepath.Base(hostPath), "devices")
	deviceFiles, err := ioutil.ReadDir(iommuDevicesPath)
	if err != nil {
		return false, err
	}
	// Pass all devices in iommu group
	for _, deviceFile := range deviceFiles {
		tokens := strings.Split(deviceFile.Name(), ":")
		vfioDeviceType := config.VFIODeviceErrorType
		if len(tokens) == 3 {
			vfioDeviceType = config.VFIODeviceNormalType
		} else {
			tokens = strings.Split(deviceFile.Name(), "-")
			if len(tokens) == 5 {
				vfioDeviceType = config.VFIODeviceMediatedType
			}
		}
		deviceFileName := filepath.Join(iommuDevicesPath, deviceFile.Name())
		switch vfioDeviceType {
		case config.VFIODeviceNormalType:
			var buf []byte
			buf, err = ioutil.ReadFile(filepath.Join(deviceFileName, "class"))
			if err != nil {
				err = fmt.Errorf("failed to read class for %v, error:%v", deviceFileName, err)
			} else {
				pciClassCode := strings.Split(string(buf), "\n")[0]
				return strings.HasPrefix(pciClassCode, PCIClassCodeDisplayController), nil
			}
		case config.VFIODeviceMediatedType:
			err = fmt.Errorf("Get pci class code for VFIODeviceMediatedType is not yet supported (%v)", deviceFileName)
		default:
			err = fmt.Errorf("Incorrect tokens found while check vfio class code: %s", deviceFileName)
		}
	}
	return false, err
}
