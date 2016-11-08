//
//Copyright [2016] [SnapRoute Inc]
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
//	 Unless required by applicable law or agreed to in writing, software
//	 distributed under the License is distributed on an "AS IS" BASIS,
//	 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//	 See the License for the specific language governing permissions and
//	 limitations under the License.
//
// _______  __       __________   ___      _______.____    __    ____  __  .___________.  ______  __    __
// |   ____||  |     |   ____\  \ /  /     /       |\   \  /  \  /   / |  | |           | /      ||  |  |  |
// |  |__   |  |     |  |__   \  V  /     |   (----` \   \/    \/   /  |  | `---|  |----`|  ,----'|  |__|  |
// |   __|  |  |     |   __|   >   <       \   \      \            /   |  |     |  |     |  |     |   __   |
// |  |     |  `----.|  |____ /  .  \  .----)   |      \    /\    /    |  |     |  |     |  `----.|  |  |  |
// |__|     |_______||_______/__/ \__\ |_______/        \__/  \__/     |__|     |__|      \______||__|  |__|
//
package server

import (
	"l3/vrrp/config"
	"l3/vrrp/debug"
)

func (svr *VrrpServer) populateState(key KeyInfo) *config.State {
	entry := config.State{}
	intf, exists := svr.Intf[key]
	if !exists {
		debug.Logger.Err("No vrrp interface configured for:", key)
		return nil
	}
	intf.Fsm.GetStateInfo(&entry)
	return &entry
}

func (svr *VrrpServer) GetV4Intfs(idx, cnt int) (int, int, []config.State) {
	var nextIdx int
	var count int
	var i, j int

	length := len(svr.v4Intfs)
	if length == 0 {
		debug.Logger.Err("No Vrrp V4 or version 2 vrrp configured")
		return 0, 0, nil
	}
	var result []config.State

	for i, j = 0, idx; i < cnt && j < length; j++ {
		key := svr.v4Intfs[j]
		rv := svr.populateState(key)
		result = append(result, *rv)
		i++
	}
	if j == length {
		nextIdx = 0
	}
	count = i
	return nextIdx, count, result
}

func (svr *VrrpServer) GetV6Intfs(idx, cnt int) (int, int, []config.State) {
	var nextIdx int
	var count int
	var i, j int

	length := len(svr.v6Intfs)
	if length == 0 {
		debug.Logger.Err("No Vrrp V6 or version 3 vrrp configured")
		return 0, 0, nil
	}
	var result []config.State

	for i, j = 0, idx; i < cnt && j < length; j++ {
		key := svr.v6Intfs[j]
		rv := svr.populateState(key)
		result = append(result, *rv)
		i++
	}
	if j == length {
		nextIdx = 0
	}
	count = i
	return nextIdx, count, result
}

func (svr *VrrpServer) GetEntry(intfRef string, vrid int32, version uint8) *config.State {
	key := KeyInfo{intfRef, vrid, version}
	return svr.populateState(key)
}
