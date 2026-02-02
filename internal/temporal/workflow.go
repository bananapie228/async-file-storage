package temporal

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// DownloadWorkflow orchestrates the file downloading process.
// Now it takes requestID to track progress in the database.
func DownloadWorkflow(ctx workflow.Context, requestID int, urls []string) ([]string, error) {
	// Define Activity options: timeout and retry policy
	options := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, options)

	logger := workflow.GetLogger(ctx)
	logger.Info("Workflow started", "RequestID", requestID, "URLCount", len(urls))

	// Define our activities container
	var a *Activities

	// Execute the downloading activity
	var results []string
	err := workflow.ExecuteActivity(ctx, a.DownloadFilesActivity, requestID, urls).Get(ctx, &results)

	if err != nil {
		logger.Error("Workflow failed", "Error", err)
		return nil, err
	}

	logger.Info("Workflow completed successfully", "RequestID", requestID)
	return results, nil
}
