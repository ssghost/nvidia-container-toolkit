/**
# Copyright 2024 NVIDIA CORPORATION
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
**/

package dgpu

import (
	"testing"

	"github.com/NVIDIA/go-nvlib/pkg/nvlib/device"
	"github.com/NVIDIA/go-nvml/pkg/nvml"
	"github.com/NVIDIA/go-nvml/pkg/nvml/mock"
	testlog "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"

	"github.com/NVIDIA/nvidia-container-toolkit/internal/discover"
)

// TODO: In order to properly test this, we need a mechanism to inject /
// override the char device discoverer.
func TestNewNvmlDGPUDiscoverer(t *testing.T) {
	logger, _ := testlog.NewNullLogger()

	nvmllib := &mock.Interface{}
	devicelib := device.New(
		nvmllib,
	)

	testCases := []struct {
		description     string
		device          nvml.Device
		expectedError   error
		expectedDevices []discover.Device
		expectedHooks   []discover.Hook
		expectedMounts  []discover.Mount
	}{
		{
			description: "",
			device: &mock.Device{
				GetMinorNumberFunc: func() (int, nvml.Return) {
					return 3, nvml.SUCCESS
				},
				GetPciInfoFunc: func() (nvml.PciInfo, nvml.Return) {
					var busID [32]int8
					for i, b := range []byte("00000000:45:00:00") {
						busID[i] = int8(b)
					}
					info := nvml.PciInfo{
						BusId: busID,
					}
					return info, nvml.SUCCESS
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			o := &options{logger: logger}

			device, err := devicelib.NewDevice(tc.device)
			require.NoError(t, err)

			d, err := o.newNvmlDGPUDiscoverer(&toRequiredInfo{device})
			require.ErrorIs(t, err, tc.expectedError)

			devices, _ := d.Devices()
			require.EqualValues(t, tc.expectedDevices, devices)
			hooks, _ := d.Hooks()
			require.EqualValues(t, tc.expectedHooks, hooks)
			mounts, _ := d.Mounts()
			require.EqualValues(t, tc.expectedMounts, mounts)
		})
	}
}
