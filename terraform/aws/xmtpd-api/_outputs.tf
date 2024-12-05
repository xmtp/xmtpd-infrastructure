output "load_balancer_address" {
  description = "The full address for the load balancer"
  value       = aws_lb.public.dns_name
}

output "load_balancer_port" {
  description = "The port for the load balancer"
  value       = local.public_port
}

output "load_balancer_zone_id" {
  description = "The zone ID for the load balancer"
  value       = aws_lb.public.zone_id
}
