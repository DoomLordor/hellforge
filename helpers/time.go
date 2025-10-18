package helpers

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func Now() time.Time {
	return time.Now().UTC().Truncate(time.Millisecond)
}

func NowPointer() *time.Time {
	return ValueToPtr(Now())
}

// TimePointerToProto converted go *time.Time to proto *timestamp
func TimePointerToProto(t *time.Time) *timestamppb.Timestamp {
	if t == nil {
		return nil
	}
	return timestamppb.New(*t)
}

// TimeProtoToPointer convert proto *timestamp to go *time.Time
func TimeProtoToPointer(t *timestamppb.Timestamp) *time.Time {
	if t == nil {
		return nil
	}
	result := t.AsTime()
	return &result
}

// GetCurrentTimeInTZ return current time in required time zone
func GetCurrentTimeInTZ(tz string) time.Time {
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return time.Now().UTC()
	}

	return time.Now().In(loc)
}

// GetTimeInTZ return time in required time zone
func GetTimeInTZ(now time.Time, tz string) time.Time {
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return now.UTC()
	}

	return now.In(loc)
}
