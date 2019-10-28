package template

import (
	"github.com/awslabs/goformation/v3/cloudformation"
	"github.com/awslabs/goformation/v3/cloudformation/cloudwatch"
)

func ValidateAWSCloudWatchAlarm(
	projectName, configName, resourceName string,
	template *cloudformation.Template,
	res *cloudwatch.Alarm,
) error {
	if res.AlarmName != "" {
		return resourceError(res, resourceName, "Names are overwritten")
	}

	res.AlarmName = normalizeName("fenrir", projectName, configName, resourceName, 255)

	return nil
}
