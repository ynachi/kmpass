package app

import "testing"

func TestCluster_validateConfig(t *testing.T) {
	type fields struct {
		Name              string
		PublicAPIEndpoint string
		PodSubnet         string
		CmpNodesIPs       []string
		CmpNodesMemory    string
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cluster := &Cluster{
				Name:              tt.fields.Name,
				PublicAPIEndpoint: tt.fields.PublicAPIEndpoint,
				PodSubnet:         tt.fields.PodSubnet,
				CmpNodesIPs:       tt.fields.CmpNodesIPs,
				CmpNodesMemory:    tt.fields.CmpNodesMemory,
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
