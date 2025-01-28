package tf

import (
	"github.com/gruntwork-io/terratest/modules/testing"
	"github.com/gruntwork-io/terratest/modules/logger"
	"strings"
	"github.com/stretchr/testify/require"
)


func GetValue(t testing.TestingT, outputs map[string]interface{}, key string) (interface{}) {
    output, ok := outputs[key].(map[string]interface{})
    require.True(t, ok, "Error finding key %s in JSON %s", key, outputs)
    value, exists := output["value"]
    require.True(t, exists, "Error finding value %s from JSON %s Error", key, outputs)
    // Return the value as is - it could be a string, list, or any other type
    return value
}

func GetStringValue(t testing.TestingT, outputs map[string]interface{}, key string) (string) {
    value := GetValue(t, outputs, key)
    strValue, ok := value.(string)
    require.True(t, ok, "Fond value %s for %s is not a string", value, key)  
    return strValue
}

func GetStringListValue(t testing.TestingT, outputs map[string]interface{}, key string) ([]string) {
    value := GetValue(t, outputs, key)

    if listValue, ok := value.([]interface{}); ok {
        result := make([]string, len(listValue))
        for i, v := range listValue {
            if strValue, ok := v.(string); ok {
                result[i] = strValue
            } else {
	        logger.Logf(t, "value at index %d for key %s is not a string", i, key)
                return nil
            }
        }
        return result
    }
    logger.Logf(t, "value for key %s is not a list", key)
    return nil
}

func HasKeyWithPrefix(t testing.TestingT, outputs map[string]interface{}, prefix string) (bool) {

    for key := range outputs {
        if strings.HasPrefix(key, prefix) {
            return true
        }
    }

    return false
}


