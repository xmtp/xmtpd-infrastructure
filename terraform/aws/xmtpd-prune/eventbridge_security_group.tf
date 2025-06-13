resource "aws_iam_role" "eventbridge_invoke_ecs" {
  name_prefix = "eventbridgeEcsInvokeRole"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Principal = {
        Service = "events.amazonaws.com"
      }
      Action = "sts:AssumeRole"
    }]
  })
}

resource "aws_iam_policy" "invoke_ecs_policy" {
  name_prefix = "InvokeECSTask"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "ecs:RunTask"
        ]
        Resource = "*"
      },
      {
        Effect = "Allow"
        Action = [
          "iam:PassRole"
        ]
        Resource = [
          module.task_definition_prune.execution_role_arn,
          module.task_definition_prune.task_role_arn
        ],
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "invoke_ecs_policy_attachment" {
  role       = aws_iam_role.eventbridge_invoke_ecs.name
  policy_arn = aws_iam_policy.invoke_ecs_policy.arn
}