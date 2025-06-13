output "task_definition_arn" {
  description = "The ARN of the task definition"
  value       = aws_ecs_task_definition.task.arn
}

output "execution_role_arn" {
  value = aws_iam_role.execution_role.arn
}

output "task_role_arn" {
  value = aws_iam_role.task_role.arn
}