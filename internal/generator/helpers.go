package generator

import (
	"fmt"
	"strings"

	"custodian/pkg/config"
)

// regionToString converts a Region enum to its string representation
func regionToString(r config.Region) string {
	regionMap := map[config.Region]string{
		config.Region_REGION_US_CENTRAL1:     "us-central1",
		config.Region_REGION_US_EAST1:        "us-east1",
		config.Region_REGION_US_EAST4:        "us-east4",
		config.Region_REGION_US_WEST1:        "us-west1",
		config.Region_REGION_US_WEST2:        "us-west2",
		config.Region_REGION_US_WEST3:        "us-west3",
		config.Region_REGION_US_WEST4:        "us-west4",
		config.Region_REGION_EUROPE_WEST1:    "europe-west1",
		config.Region_REGION_EUROPE_WEST2:    "europe-west2",
		config.Region_REGION_EUROPE_WEST3:    "europe-west3",
		config.Region_REGION_EUROPE_WEST4:    "europe-west4",
		config.Region_REGION_EUROPE_WEST6:    "europe-west6",
		config.Region_REGION_EUROPE_NORTH1:   "europe-north1",
		config.Region_REGION_ASIA_EAST1:      "asia-east1",
		config.Region_REGION_ASIA_EAST2:      "asia-east2",
		config.Region_REGION_ASIA_NORTHEAST1: "asia-northeast1",
		config.Region_REGION_ASIA_NORTHEAST2: "asia-northeast2",
		config.Region_REGION_ASIA_NORTHEAST3: "asia-northeast3",
		config.Region_REGION_ASIA_SOUTH1:     "asia-south1",
		config.Region_REGION_ASIA_SOUTHEAST1: "asia-southeast1",
		config.Region_REGION_ASIA_SOUTHEAST2: "asia-southeast2",
	}
	
	if str, ok := regionMap[r]; ok {
		return str
	}
	return "us-central1" // default
}

// zoneToString converts a Zone enum to its string representation
func zoneToString(z config.Zone) string {
	zoneMap := map[config.Zone]string{
		config.Zone_ZONE_US_CENTRAL1_A: "us-central1-a",
		config.Zone_ZONE_US_CENTRAL1_B: "us-central1-b",
		config.Zone_ZONE_US_CENTRAL1_C: "us-central1-c",
		config.Zone_ZONE_US_CENTRAL1_F: "us-central1-f",
		config.Zone_ZONE_US_EAST1_B:    "us-east1-b",
		config.Zone_ZONE_US_EAST1_C:    "us-east1-c",
		config.Zone_ZONE_US_EAST1_D:    "us-east1-d",
		config.Zone_ZONE_US_EAST4_A:    "us-east4-a",
		config.Zone_ZONE_US_EAST4_B:    "us-east4-b",
		config.Zone_ZONE_US_EAST4_C:    "us-east4-c",
		config.Zone_ZONE_US_WEST1_A:    "us-west1-a",
		config.Zone_ZONE_US_WEST1_B:    "us-west1-b",
		config.Zone_ZONE_US_WEST1_C:    "us-west1-c",
		config.Zone_ZONE_US_WEST2_A:    "us-west2-a",
		config.Zone_ZONE_US_WEST2_B:    "us-west2-b",
		config.Zone_ZONE_US_WEST2_C:    "us-west2-c",
		config.Zone_ZONE_EUROPE_WEST1_B: "europe-west1-b",
		config.Zone_ZONE_EUROPE_WEST1_C: "europe-west1-c",
		config.Zone_ZONE_EUROPE_WEST1_D: "europe-west1-d",
		config.Zone_ZONE_ASIA_EAST1_A:   "asia-east1-a",
		config.Zone_ZONE_ASIA_EAST1_B:   "asia-east1-b",
		config.Zone_ZONE_ASIA_EAST1_C:   "asia-east1-c",
	}
	
	if str, ok := zoneMap[z]; ok {
		return str
	}
	return "us-central1-a" // default
}

// machineTypeToString converts a MachineType enum to its string representation
func machineTypeToString(mt config.MachineType) string {
	machineTypeMap := map[config.MachineType]string{
		config.MachineType_MACHINE_E2_MICRO:       "e2-micro",
		config.MachineType_MACHINE_E2_SMALL:       "e2-small",
		config.MachineType_MACHINE_E2_MEDIUM:      "e2-medium",
		config.MachineType_MACHINE_E2_STANDARD_2:  "e2-standard-2",
		config.MachineType_MACHINE_E2_STANDARD_4:  "e2-standard-4",
		config.MachineType_MACHINE_E2_STANDARD_8:  "e2-standard-8",
		config.MachineType_MACHINE_E2_STANDARD_16: "e2-standard-16",
		config.MachineType_MACHINE_N1_STANDARD_1:  "n1-standard-1",
		config.MachineType_MACHINE_N1_STANDARD_2:  "n1-standard-2",
		config.MachineType_MACHINE_N1_STANDARD_4:  "n1-standard-4",
		config.MachineType_MACHINE_N1_STANDARD_8:  "n1-standard-8",
		config.MachineType_MACHINE_N1_STANDARD_16: "n1-standard-16",
		config.MachineType_MACHINE_N2_STANDARD_2:  "n2-standard-2",
		config.MachineType_MACHINE_N2_STANDARD_4:  "n2-standard-4",
		config.MachineType_MACHINE_N2_STANDARD_8:  "n2-standard-8",
		config.MachineType_MACHINE_N2_STANDARD_16: "n2-standard-16",
		config.MachineType_MACHINE_C2_STANDARD_4:  "c2-standard-4",
		config.MachineType_MACHINE_C2_STANDARD_8:  "c2-standard-8",
		config.MachineType_MACHINE_C2_STANDARD_16: "c2-standard-16",
	}
	
	if str, ok := machineTypeMap[mt]; ok {
		return str
	}
	return "e2-medium" // default
}

// apiToString converts a GcpApi enum to its service name
func apiToString(api config.GcpApi) string {
	apiMap := map[config.GcpApi]string{
		config.GcpApi_API_COMPUTE:            "compute.googleapis.com",
		config.GcpApi_API_CONTAINER:          "container.googleapis.com",
		config.GcpApi_API_SQL_ADMIN:          "sqladmin.googleapis.com",
		config.GcpApi_API_STORAGE:            "storage.googleapis.com",
		config.GcpApi_API_BIGQUERY:           "bigquery.googleapis.com",
		config.GcpApi_API_PUBSUB:             "pubsub.googleapis.com",
		config.GcpApi_API_DATAFLOW:           "dataflow.googleapis.com",
		config.GcpApi_API_MONITORING:         "monitoring.googleapis.com",
		config.GcpApi_API_LOGGING:            "logging.googleapis.com",
		config.GcpApi_API_IAM:                "iam.googleapis.com",
		config.GcpApi_API_RESOURCE_MANAGER:   "cloudresourcemanager.googleapis.com",
		config.GcpApi_API_CLOUD_BUILD:        "cloudbuild.googleapis.com",
		config.GcpApi_API_CLOUD_FUNCTIONS:    "cloudfunctions.googleapis.com",
		config.GcpApi_API_CLOUD_RUN:          "run.googleapis.com",
		config.GcpApi_API_KUBERNETES_ENGINE:  "container.googleapis.com",
		config.GcpApi_API_CLOUD_DNS:          "dns.googleapis.com",
		config.GcpApi_API_CLOUD_CDN:          "compute.googleapis.com",
		config.GcpApi_API_LOAD_BALANCING:     "compute.googleapis.com",
		config.GcpApi_API_VPC_ACCESS:         "vpcaccess.googleapis.com",
		config.GcpApi_API_FIREWALL:           "compute.googleapis.com",
	}
	
	if str, ok := apiMap[api]; ok {
		return str
	}
	return ""
}

// indent adds indentation to each line of the input string
func indent(spaces int, text string) string {
	indentation := strings.Repeat(" ", spaces)
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if strings.TrimSpace(line) != "" {
			lines[i] = indentation + line
		}
	}
	return strings.Join(lines, "\n")
}

// quote wraps a string in double quotes
func quote(s string) string {
	return fmt.Sprintf(`"%s"`, s)
}