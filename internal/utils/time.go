package utils

import "time"

func ParseDatePtr(dateStr *string) *time.Time {
    if dateStr == nil || *dateStr == "" {
        return nil
    }
    
    t, err := time.Parse(time.RFC3339, *dateStr)
    if err != nil {
        return nil
    }
    
    return &t
}