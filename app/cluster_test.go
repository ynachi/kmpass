package app

import (
	"testing"
)

func TestCluster_ValidateConfig(t *testing.T) {
	type fields struct {
		Name              string
		PodSubnet         string
		CmpNodesMemory    string
		CmpNodesCores     int
		CmpNodesNumber    int
		CmpNodesDiskSize  string
		CtrlNodesMemory   string
		CtrlNodesCores    int
		CtrlNodesNumber   int
		CtrlNodesDiskSize string
		LBNodeMemory      string
		LBNodeCore        int
		LBNodeDiskSize    string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "cluster_good_1",
			fields: fields{
				Name:              "cluster100",
				PodSubnet:         "10.10.10.0/24",
				CmpNodesMemory:    "4G",
				CmpNodesCores:     3,
				CmpNodesDiskSize:  "10G",
				CtrlNodesMemory:   "2G",
				CtrlNodesCores:    3,
				CtrlNodesDiskSize: "20G",
				LBNodeMemory:      "2G",
				LBNodeCore:        2,
				LBNodeDiskSize:    "10G",
				CmpNodesNumber:    1,
				CtrlNodesNumber:   3,
			},
			wantErr: false,
		},
		{
			name: "cluster_bad_pod_subnet",
			fields: fields{
				Name:              "cluster201",
				PodSubnet:         "10.10.10/24",
				CmpNodesMemory:    "4G",
				CmpNodesCores:     3,
				CmpNodesDiskSize:  "10G",
				CtrlNodesMemory:   "2G",
				CtrlNodesCores:    3,
				CtrlNodesDiskSize: "20G",
				LBNodeMemory:      "2G",
				LBNodeCore:        2,
				LBNodeDiskSize:    "10G",
				CtrlNodesNumber:   3,
				CmpNodesNumber:    5,
			},
			wantErr: true,
		},
		{
			name: "cluster_bad_memory_fmt",
			fields: fields{
				Name:              "cluster100",
				PodSubnet:         "10.10.10.0/24",
				CmpNodesMemory:    "4Gg",
				CmpNodesCores:     3,
				CmpNodesDiskSize:  "10G",
				CtrlNodesMemory:   "2G",
				CtrlNodesCores:    3,
				CtrlNodesDiskSize: "20G",
				LBNodeMemory:      "2G",
				LBNodeCore:        2,
				LBNodeDiskSize:    "10G",
				CmpNodesNumber:    1,
				CtrlNodesNumber:   5,
			},
			wantErr: true,
		},
		{
			name: "cluster_bad_even_ctrl_nodes",
			fields: fields{
				Name:              "cluster100",
				PodSubnet:         "10.10.10.0/24",
				CmpNodesMemory:    "4g",
				CmpNodesCores:     3,
				CmpNodesDiskSize:  "10G",
				CtrlNodesMemory:   "2G",
				CtrlNodesCores:    3,
				CtrlNodesDiskSize: "20G",
				LBNodeMemory:      "2G",
				LBNodeCore:        2,
				LBNodeDiskSize:    "10G",
				CmpNodesNumber:    1,
				CtrlNodesNumber:   6,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cluster := &Cluster{
				Name:              tt.fields.Name,
				PodSubnet:         tt.fields.PodSubnet,
				CmpNodesMemory:    tt.fields.CmpNodesMemory,
				CmpNodesCores:     tt.fields.CmpNodesCores,
				CmpNodesNumber:    tt.fields.CmpNodesNumber,
				CmpNodesDiskSize:  tt.fields.CmpNodesDiskSize,
				CtrlNodesMemory:   tt.fields.CtrlNodesMemory,
				CtrlNodesCores:    tt.fields.CtrlNodesCores,
				CtrlNodesNumber:   tt.fields.CtrlNodesNumber,
				CtrlNodesDiskSize: tt.fields.CtrlNodesDiskSize,
				LBNodeMemory:      tt.fields.LBNodeMemory,
				LBNodeCore:        tt.fields.LBNodeCore,
				LBNodeDiskSize:    tt.fields.LBNodeDiskSize,
			}
			if err := cluster.ValidateConfig(); (err != nil) != tt.wantErr {
				t.Errorf("ValidateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
