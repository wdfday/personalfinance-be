package domain

// SnapshotType represents the type of portfolio snapshot
type SnapshotType string

const (
	SnapshotTypeManual    SnapshotType = "manual"    // Manually created snapshot
	SnapshotTypeAutomatic SnapshotType = "automatic" // Automatically created snapshot
	SnapshotTypePeriodic  SnapshotType = "periodic"  // Periodic snapshot (daily, weekly, monthly)
)

// IsValid checks if the snapshot type is valid
func (st SnapshotType) IsValid() bool {
	switch st {
	case SnapshotTypeManual, SnapshotTypeAutomatic, SnapshotTypePeriodic:
		return true
	}
	return false
}

// String returns the string representation
func (st SnapshotType) String() string {
	return string(st)
}

// SnapshotPeriod represents the period for periodic snapshots
type SnapshotPeriod string

const (
	SnapshotPeriodDaily   SnapshotPeriod = "daily"
	SnapshotPeriodWeekly  SnapshotPeriod = "weekly"
	SnapshotPeriodMonthly SnapshotPeriod = "monthly"
	SnapshotPeriodYearly  SnapshotPeriod = "yearly"
)

// IsValid checks if the snapshot period is valid
func (sp SnapshotPeriod) IsValid() bool {
	switch sp {
	case SnapshotPeriodDaily, SnapshotPeriodWeekly, SnapshotPeriodMonthly, SnapshotPeriodYearly:
		return true
	}
	return false
}

// String returns the string representation
func (sp SnapshotPeriod) String() string {
	return string(sp)
}
