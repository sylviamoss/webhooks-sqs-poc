resource "aws_sqs_queue" "webhooks_sqs" {
  name                      = "webhooks-queue"
  message_retention_seconds = 432000 # (5 days)
  receive_wait_time_seconds = 20
  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.webhooks_sqs_deadletter.arn
    maxReceiveCount     = 5
  })
  tags = {
    Reason = "moss-webhooks-testing"
  }
}

resource "aws_sqs_queue" "webhooks_sqs_deadletter" {
  name                      = "webhooks-deadletter-queue"
  message_retention_seconds = 1209600 # (14 days)
  tags = {
    Reason = "moss-webhooks-testing"
  }
}