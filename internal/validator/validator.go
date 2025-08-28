package validator

import (
	"fmt"
	"net"
	"regexp"

	"custoodian/pkg/config"

	"github.com/bufbuild/protovalidate-go"
)

// ValidateConfig validates a complete configuration
func ValidateConfig(cfg *config.Config) error {
	// First, validate using protovalidate constraints
	validator, err := protovalidate.New()
	if err != nil {
		return fmt.Errorf("failed to create validator: %w", err)
	}

	if err := validator.Validate(cfg); err != nil {
		return fmt.Errorf("proto validation failed: %w", err)
	}

	// Custom business logic validations
	if err := validateProject(cfg.Project); err != nil {
		return fmt.Errorf("project validation failed: %w", err)
	}

	if cfg.Networking != nil {
		if err := validateNetworking(cfg.Networking); err != nil {
			return fmt.Errorf("networking validation failed: %w", err)
		}
	}

	if cfg.Compute != nil {
		if err := validateCompute(cfg.Compute); err != nil {
			return fmt.Errorf("compute validation failed: %w", err)
		}
	}

	if len(cfg.LoadBalancers) > 0 {
		if err := validateLoadBalancers(cfg.LoadBalancers); err != nil {
			return fmt.Errorf("load balancer validation failed: %w", err)
		}
	}

	if cfg.Iam != nil {
		if err := validateIAM(cfg.Iam); err != nil {
			return fmt.Errorf("IAM validation failed: %w", err)
		}
	}

	if cfg.Storage != nil {
		if err := validateStorage(cfg.Storage); err != nil {
			return fmt.Errorf("storage validation failed: %w", err)
		}
	}

	// Cross-resource validations
	if err := validateCrossReferences(cfg); err != nil {
		return fmt.Errorf("cross-reference validation failed: %w", err)
	}

	return nil
}

// validateProject validates project configuration
func validateProject(project *config.Project) error {
	if project == nil {
		return fmt.Errorf("project configuration is required")
	}

	// Validate project ID format (GCP-specific rules)
	if !isValidGCPProjectID(project.Id) {
		return fmt.Errorf("invalid project ID: %s (must be 6-30 characters, lowercase letters, numbers, and hyphens, start with letter, end with letter or number)", project.Id)
	}

	// Validate billing account format
	if project.BillingAccount != "" && !isValidBillingAccount(project.BillingAccount) {
		return fmt.Errorf("invalid billing account format: %s", project.BillingAccount)
	}

	// Validate that organization_id and folder_id are mutually exclusive
	if project.OrganizationId != "" && project.FolderId != "" {
		return fmt.Errorf("organization_id and folder_id are mutually exclusive")
	}

	return nil
}

// validateNetworking validates networking configuration
func validateNetworking(networking *config.Networking) error {
	// Validate reserved IPs
	for _, ip := range networking.ReservedIps {
		if err := validateReservedIP(ip); err != nil {
			return fmt.Errorf("invalid reserved IP %s: %w", ip.Name, err)
		}
	}

	// Validate VPCs
	for _, vpc := range networking.Vpcs {
		if err := validateVPC(vpc); err != nil {
			return fmt.Errorf("invalid VPC %s: %w", vpc.Name, err)
		}
	}

	// Validate firewall rules
	for _, rule := range networking.FirewallRules {
		if err := validateFirewallRule(rule); err != nil {
			return fmt.Errorf("invalid firewall rule %s: %w", rule.Name, err)
		}
	}

	// Validate NAT gateways
	for _, nat := range networking.NatGateways {
		if err := validateNATGateway(nat); err != nil {
			return fmt.Errorf("invalid NAT gateway %s: %w", nat.Name, err)
		}
	}

	return nil
}

// validateReservedIP validates a reserved IP configuration
func validateReservedIP(ip *config.ReservedIp) error {
	// Regional IPs must have a region specified
	if ip.Type == config.ReservedIpType_RESERVED_IP_TYPE_REGIONAL && ip.Region == config.Region_REGION_UNSPECIFIED {
		return fmt.Errorf("regional reserved IP must specify a region")
	}

	// Global IPs should not have a region
	if ip.Type == config.ReservedIpType_RESERVED_IP_TYPE_GLOBAL && ip.Region != config.Region_REGION_UNSPECIFIED {
		return fmt.Errorf("global reserved IP should not specify a region")
	}

	return nil
}

// validateVPC validates a VPC configuration
func validateVPC(vpc *config.Vpc) error {
	// Validate subnets
	usedCIDRs := make(map[string]bool)
	
	for _, subnet := range vpc.Subnets {
		if err := validateSubnet(subnet); err != nil {
			return fmt.Errorf("invalid subnet %s: %w", subnet.Name, err)
		}

		// Check for CIDR overlaps
		if usedCIDRs[subnet.Cidr] {
			return fmt.Errorf("duplicate CIDR range %s in subnet %s", subnet.Cidr, subnet.Name)
		}
		usedCIDRs[subnet.Cidr] = true

		// Validate CIDR overlaps (basic check)
		for existingCIDR := range usedCIDRs {
			if existingCIDR != subnet.Cidr && cidrsOverlap(subnet.Cidr, existingCIDR) {
				return fmt.Errorf("CIDR range %s in subnet %s overlaps with existing range %s", subnet.Cidr, subnet.Name, existingCIDR)
			}
		}
	}

	return nil
}

// validateSubnet validates a subnet configuration
func validateSubnet(subnet *config.Subnet) error {
	// Validate CIDR format
	if !isValidCIDR(subnet.Cidr) {
		return fmt.Errorf("invalid CIDR format: %s", subnet.Cidr)
	}

	// Validate secondary ranges
	usedSecondaryRanges := make(map[string]bool)
	for _, secondary := range subnet.SecondaryRanges {
		if !isValidCIDR(secondary.IpCidrRange) {
			return fmt.Errorf("invalid secondary CIDR format: %s", secondary.IpCidrRange)
		}

		if usedSecondaryRanges[secondary.RangeName] {
			return fmt.Errorf("duplicate secondary range name: %s", secondary.RangeName)
		}
		usedSecondaryRanges[secondary.RangeName] = true
	}

	return nil
}

// validateFirewallRule validates a firewall rule
func validateFirewallRule(rule *config.FirewallRule) error {
	// Validate direction-specific fields
	if rule.Direction == "INGRESS" && len(rule.DestinationRanges) > 0 {
		return fmt.Errorf("INGRESS rules cannot have destination_ranges")
	}
	
	if rule.Direction == "EGRESS" && len(rule.SourceRanges) > 0 {
		return fmt.Errorf("EGRESS rules cannot have source_ranges")
	}

	if rule.Direction == "EGRESS" && len(rule.SourceTags) > 0 {
		return fmt.Errorf("EGRESS rules cannot have source_tags")
	}

	// Validate that either allow or deny is specified, but not both
	if len(rule.Allow) > 0 && len(rule.Deny) > 0 {
		return fmt.Errorf("firewall rule cannot have both allow and deny blocks")
	}

	if len(rule.Allow) == 0 && len(rule.Deny) == 0 {
		return fmt.Errorf("firewall rule must have either allow or deny block")
	}

	// Validate IP ranges
	for _, cidr := range rule.SourceRanges {
		if !isValidCIDR(cidr) {
			return fmt.Errorf("invalid source range CIDR: %s", cidr)
		}
	}

	for _, cidr := range rule.DestinationRanges {
		if !isValidCIDR(cidr) {
			return fmt.Errorf("invalid destination range CIDR: %s", cidr)
		}
	}

	return nil
}

// validateNATGateway validates a NAT gateway configuration
func validateNATGateway(nat *config.NatGateway) error {
	// Validate NAT IP allocation options
	validOptions := map[string]bool{
		"MANUAL_ONLY": true,
		"AUTO_ONLY":   true,
	}

	if !validOptions[nat.NatIpAllocateOption] {
		return fmt.Errorf("invalid NAT IP allocate option: %s", nat.NatIpAllocateOption)
	}

	// If MANUAL_ONLY, must have NAT IPs specified
	if nat.NatIpAllocateOption == "MANUAL_ONLY" && len(nat.NatIps) == 0 {
		return fmt.Errorf("MANUAL_ONLY NAT IP allocation requires nat_ips to be specified")
	}

	return nil
}

// validateCompute validates compute configuration
func validateCompute(compute *config.Compute) error {
	// Validate instance templates
	templateNames := make(map[string]bool)
	for _, template := range compute.InstanceTemplates {
		if templateNames[template.Name] {
			return fmt.Errorf("duplicate instance template name: %s", template.Name)
		}
		templateNames[template.Name] = true

		if err := validateInstanceTemplate(template); err != nil {
			return fmt.Errorf("invalid instance template %s: %w", template.Name, err)
		}
	}

	// Validate instance groups
	for _, group := range compute.InstanceGroups {
		if err := validateInstanceGroup(group); err != nil {
			return fmt.Errorf("invalid instance group %s: %w", group.Name, err)
		}

		// Check that referenced template exists
		if !templateNames[group.Template] {
			return fmt.Errorf("instance group %s references unknown template: %s", group.Name, group.Template)
		}
	}

	return nil
}

// validateInstanceTemplate validates an instance template
func validateInstanceTemplate(template *config.InstanceTemplate) error {
	// Validate disk size
	if template.DiskSizeGb < 10 {
		return fmt.Errorf("disk size must be at least 10 GB")
	}

	// Validate network interfaces
	for _, iface := range template.NetworkInterfaces {
		if iface.Network == "" && iface.Subnetwork == "" {
			return fmt.Errorf("network interface must specify either network or subnetwork")
		}
	}

	return nil
}

// validateInstanceGroup validates an instance group
func validateInstanceGroup(group *config.InstanceGroup) error {
	// Validate auto scaling configuration
	if group.AutoScaling != nil {
		if group.AutoScaling.Min > group.AutoScaling.Max {
			return fmt.Errorf("auto scaling min (%d) cannot be greater than max (%d)", group.AutoScaling.Min, group.AutoScaling.Max)
		}

		if group.AutoScaling.CpuTarget <= 0 || group.AutoScaling.CpuTarget > 1 {
			return fmt.Errorf("CPU target must be between 0 and 1, got %f", group.AutoScaling.CpuTarget)
		}
	}

	return nil
}

// validateLoadBalancers validates load balancer configurations
func validateLoadBalancers(lbs []*config.LoadBalancer) error {
	for _, lb := range lbs {
		if err := validateLoadBalancer(lb); err != nil {
			return fmt.Errorf("invalid load balancer %s: %w", lb.Name, err)
		}
	}
	return nil
}

// validateLoadBalancer validates a single load balancer
func validateLoadBalancer(lb *config.LoadBalancer) error {
	// Validate health check if present
	if lb.HealthCheck != nil {
		if err := validateHealthCheck(lb.HealthCheck); err != nil {
			return fmt.Errorf("invalid health check: %w", err)
		}
	}

	return nil
}

// validateHealthCheck validates a health check configuration
func validateHealthCheck(hc *config.HealthCheck) error {
	// Validate port range
	if hc.Port <= 0 || hc.Port > 65535 {
		return fmt.Errorf("invalid port: %d", hc.Port)
	}

	// Validate timeouts
	if hc.TimeoutSec >= hc.CheckIntervalSec {
		return fmt.Errorf("timeout_sec (%d) must be less than check_interval_sec (%d)", hc.TimeoutSec, hc.CheckIntervalSec)
	}

	return nil
}

// validateIAM validates IAM configuration
func validateIAM(iam *config.Iam) error {
	// Validate service accounts
	accountIds := make(map[string]bool)
	for _, sa := range iam.ServiceAccounts {
		if accountIds[sa.AccountId] {
			return fmt.Errorf("duplicate service account ID: %s", sa.AccountId)
		}
		accountIds[sa.AccountId] = true

		if err := validateServiceAccount(sa); err != nil {
			return fmt.Errorf("invalid service account %s: %w", sa.AccountId, err)
		}
	}

	// Validate custom roles
	roleIds := make(map[string]bool)
	for _, role := range iam.CustomRoles {
		if roleIds[role.RoleId] {
			return fmt.Errorf("duplicate custom role ID: %s", role.RoleId)
		}
		roleIds[role.RoleId] = true

		if err := validateCustomRole(role); err != nil {
			return fmt.Errorf("invalid custom role %s: %w", role.RoleId, err)
		}
	}

	return nil
}

// validateServiceAccount validates a service account configuration
func validateServiceAccount(sa *config.ServiceAccount) error {
	// Validate account ID format
	if !isValidServiceAccountId(sa.AccountId) {
		return fmt.Errorf("invalid service account ID format: %s", sa.AccountId)
	}

	return nil
}

// validateCustomRole validates a custom role configuration
func validateCustomRole(role *config.CustomRole) error {
	// Validate that permissions are not empty
	if len(role.Permissions) == 0 {
		return fmt.Errorf("custom role must have at least one permission")
	}

	// Validate stage values
	validStages := map[string]bool{
		"ALPHA":      true,
		"BETA":       true,
		"GA":         true,
		"DEPRECATED": true,
	}

	if role.Stage != "" && !validStages[role.Stage] {
		return fmt.Errorf("invalid stage: %s", role.Stage)
	}

	return nil
}

// validateStorage validates storage configuration
func validateStorage(storage *config.Storage) error {
	bucketNames := make(map[string]bool)
	
	for _, bucket := range storage.Buckets {
		if bucketNames[bucket.Name] {
			return fmt.Errorf("duplicate bucket name: %s", bucket.Name)
		}
		bucketNames[bucket.Name] = true

		if err := validateStorageBucket(bucket); err != nil {
			return fmt.Errorf("invalid storage bucket %s: %w", bucket.Name, err)
		}
	}

	return nil
}

// validateStorageBucket validates a storage bucket configuration
func validateStorageBucket(bucket *config.StorageBucket) error {
	// Validate bucket name format (GCS-specific rules)
	if !isValidBucketName(bucket.Name) {
		return fmt.Errorf("invalid bucket name format: %s", bucket.Name)
	}

	// Validate storage class
	validClasses := map[string]bool{
		"STANDARD": true,
		"NEARLINE": true,
		"COLDLINE": true,
		"ARCHIVE":  true,
	}

	if bucket.StorageClass != "" && !validClasses[bucket.StorageClass] {
		return fmt.Errorf("invalid storage class: %s", bucket.StorageClass)
	}

	return nil
}

// validateCrossReferences validates cross-resource references
func validateCrossReferences(cfg *config.Config) error {
	// Collect all resource names for validation
	resources := collectResourceNames(cfg)

	// Validate load balancer references
	for _, lb := range cfg.LoadBalancers {
		// Validate IP reference
		if lb.Ip != "" && !resources.reservedIPs[lb.Ip] {
			return fmt.Errorf("load balancer %s references unknown reserved IP: %s", lb.Name, lb.Ip)
		}

		// Validate backend reference
		if !resources.instanceGroups[lb.Backend] {
			return fmt.Errorf("load balancer %s references unknown backend: %s", lb.Name, lb.Backend)
		}
	}

	return nil
}

// resourceNames holds collections of resource names for cross-reference validation
type resourceNames struct {
	reservedIPs     map[string]bool
	networks        map[string]bool
	subnets         map[string]bool
	instanceGroups  map[string]bool
	serviceAccounts map[string]bool
}

// collectResourceNames collects all resource names from the configuration
func collectResourceNames(cfg *config.Config) *resourceNames {
	resources := &resourceNames{
		reservedIPs:     make(map[string]bool),
		networks:        make(map[string]bool),
		subnets:         make(map[string]bool),
		instanceGroups:  make(map[string]bool),
		serviceAccounts: make(map[string]bool),
	}

	// Collect networking resources
	if cfg.Networking != nil {
		for _, ip := range cfg.Networking.ReservedIps {
			resources.reservedIPs[ip.Name] = true
		}

		for _, vpc := range cfg.Networking.Vpcs {
			resources.networks[vpc.Name] = true
			for _, subnet := range vpc.Subnets {
				resources.subnets[subnet.Name] = true
			}
		}
	}

	// Collect compute resources
	if cfg.Compute != nil {
		for _, group := range cfg.Compute.InstanceGroups {
			resources.instanceGroups[group.Name] = true
		}
	}

	// Collect IAM resources
	if cfg.Iam != nil {
		for _, sa := range cfg.Iam.ServiceAccounts {
			resources.serviceAccounts[sa.AccountId] = true
		}
	}

	return resources
}

// Utility functions for validation

func isValidGCPProjectID(id string) bool {
	if len(id) < 6 || len(id) > 30 {
		return false
	}
	match, _ := regexp.MatchString(`^[a-z][a-z0-9-]*[a-z0-9]$`, id)
	return match
}

func isValidBillingAccount(account string) bool {
	match, _ := regexp.MatchString(`^[0-9]{6}-[A-Z0-9]{6}-[A-Z0-9]{6}$`, account)
	return match
}

func isValidCIDR(cidr string) bool {
	_, _, err := net.ParseCIDR(cidr)
	return err == nil
}

func cidrsOverlap(cidr1, cidr2 string) bool {
	_, net1, err1 := net.ParseCIDR(cidr1)
	_, net2, err2 := net.ParseCIDR(cidr2)
	
	if err1 != nil || err2 != nil {
		return false
	}
	
	return net1.Contains(net2.IP) || net2.Contains(net1.IP)
}

func isValidServiceAccountId(id string) bool {
	if len(id) < 6 || len(id) > 30 {
		return false
	}
	match, _ := regexp.MatchString(`^[a-z][a-z0-9-]*[a-z0-9]$`, id)
	return match
}

func isValidBucketName(name string) bool {
	if len(name) < 3 || len(name) > 63 {
		return false
	}
	// Basic validation - GCS has more complex rules
	match, _ := regexp.MatchString(`^[a-z0-9][a-z0-9\-_.]*[a-z0-9]$`, name)
	return match
}