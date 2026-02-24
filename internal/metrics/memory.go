package metrics

// instanceClassMemoryGB maps RDS instance classes to total memory in GB
// Based on AWS RDS documentation: https://aws.amazon.com/rds/instance-types/
var instanceClassMemoryGB = map[string]float64{
	// T3 instances
	"db.t3.micro":   1,
	"db.t3.small":   2,
	"db.t3.medium":  4,
	"db.t3.large":   8,
	"db.t3.xlarge":  16,
	"db.t3.2xlarge": 32,
	// T4g instances (Graviton)
	"db.t4g.micro":  1,
	"db.t4g.small":  2,
	"db.t4g.medium": 4,
	"db.t4g.large":  8,
	// R5 instances
	"db.r5.large":   16,
	"db.r5.xlarge":  32,
	"db.r5.2xlarge": 64,
	"db.r5.4xlarge": 128,
	// R6g instances (Graviton)
	"db.r6g.large":  16,
	"db.r6g.xlarge": 32,
	// M5 instances
	"db.m5.large":   8,
	"db.m5.xlarge":  16,
	"db.m5.2xlarge": 32,
	"db.m5.4xlarge": 64,
	// M6g instances (Graviton)
	"db.m6g.large":  8,
	"db.m6g.xlarge": 16,
}

// CalculateMemoryPercentage converts FreeableMemory (bytes) to percentage available
// Returns nil if instance class is not in the mapping
func CalculateMemoryPercentage(instanceClass string, freeableMemoryBytes float64) *float64 {
	totalGB, ok := instanceClassMemoryGB[instanceClass]
	if !ok {
		return nil // Unknown instance class
	}

	totalBytes := totalGB * 1024 * 1024 * 1024
	percentAvailable := (freeableMemoryBytes / totalBytes) * 100
	return &percentAvailable
}
