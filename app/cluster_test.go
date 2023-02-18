package app

import "testing"

func TestCluster_validateConfig(t *testing.T) {
	type fields struct {
		Name              string
		PublicAPIEndpoint string
		PodSubnet         string
		CmpNodesIPs       []string
		CmpNodesMemory    string
		CmpNodesCores     int
		CmpNodesDiskSize  string
		CtrlNodesIPs      []string
		CtrlNodesMemory   string
		CtrlNodesCores    int
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
				PublicAPIEndpoint: "10.10.10.1",
				PodSubnet:         "10.10.10.0/24",
				CmpNodesIPs:       []string{"10.10.10.2", "10.10.10.3", "10.10.10.4"},
				CmpNodesMemory:    "4G",
				CmpNodesCores:     3,
				CmpNodesDiskSize:  "10G",
				CtrlNodesIPs:      []string{"10.10.10.21", "10.10.10.31", "10.10.10.41"},
				CtrlNodesMemory:   "2G",
				CtrlNodesCores:    3,
				CtrlNodesDiskSize: "20G",
				LBNodeMemory:      "2G",
				LBNodeCore:        2,
				LBNodeDiskSize:    "10G",
			},
			wantErr: false,
		},
		{
			name: "cluster_bad_1",
			fields: fields{
				Name:              "cluster201",
				PublicAPIEndpoint: "10.10.1",
				PodSubnet:         "10.10.10.0/24",
				CmpNodesIPs:       []string{"10.10.10.2", "10.10.10.3", "10.10.10.4"},
				CmpNodesMemory:    "4G",
				CmpNodesCores:     3,
				CmpNodesDiskSize:  "10G",
				CtrlNodesIPs:      []string{"10.10.10.21", "10.10.10.31", "10.10.10.41"},
				CtrlNodesMemory:   "2G",
				CtrlNodesCores:    3,
				CtrlNodesDiskSize: "20G",
				LBNodeMemory:      "2G",
				LBNodeCore:        2,
				LBNodeDiskSize:    "10G",
			},
			wantErr: true,
		},
		{
			name: "cluster_bad_2",
			fields: fields{
				Name:              "cluster100",
				PublicAPIEndpoint: "10.10.10.1",
				PodSubnet:         "10.10.10.0/24",
				CmpNodesIPs:       []string{"10.10.10.2", "10.10.10.3", "10.10.10.4"},
				CmpNodesMemory:    "4Gg",
				CmpNodesCores:     3,
				CmpNodesDiskSize:  "10G",
				CtrlNodesIPs:      []string{"10.10.10.21", "10.10.10.31", "10.10.10.41"},
				CtrlNodesMemory:   "2G",
				CtrlNodesCores:    3,
				CtrlNodesDiskSize: "20G",
				LBNodeMemory:      "2G",
				LBNodeCore:        2,
				LBNodeDiskSize:    "10G",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cluster := &Cluster{
				Name:              tt.fields.Name,
				PublicAPIEndpoint: tt.fields.PublicAPIEndpoint,
				PodSubnet:         tt.fields.PodSubnet,
				CmpNodesIPs:       tt.fields.CmpNodesIPs,
				CmpNodesMemory:    tt.fields.CmpNodesMemory,
				CmpNodesCores:     tt.fields.CmpNodesCores,
				CmpNodesDiskSize:  tt.fields.CmpNodesDiskSize,
				CtrlNodesIPs:      tt.fields.CtrlNodesIPs,
				CtrlNodesMemory:   tt.fields.CtrlNodesMemory,
				CtrlNodesCores:    tt.fields.CtrlNodesCores,
				CtrlNodesDiskSize: tt.fields.CtrlNodesDiskSize,
				LBNodeMemory:      tt.fields.LBNodeMemory,
				LBNodeCore:        tt.fields.LBNodeCore,
				LBNodeDiskSize:    tt.fields.LBNodeDiskSize,
			}
			if err := cluster.validateConfig(); (err != nil) != tt.wantErr {
				t.Errorf("validateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
