import express, { NextFunction, Request, Response } from "express"
import pino from "pino"
import client from "prom-client"
import { randomUUID } from "node:crypto"
import { NodeSDK } from "@opentelemetry/sdk-node"
import { getNodeAutoInstrumentations } from "@opentelemetry/auto-instrumentations-node"
import { OTLPTraceExporter } from "@opentelemetry/exporter-trace-otlp-grpc"

const otelEndpoint = process.env.OTEL_EXPORTER_OTLP_ENDPOINT ?? "http://otel-collector:4317"
const sdk = new NodeSDK({
  traceExporter: new OTLPTraceExporter({ url: otelEndpoint }),
  instrumentations: [getNodeAutoInstrumentations()]
})
void sdk.start()

const logger = pino({ level: "info" })

const register = new client.Registry()
client.collectDefaultMetrics({ register })
const requestTotal = new client.Counter({
  name: "vedo_requests_total",
  help: "Total requests",
  labelNames: ["method", "path", "status"]
})
const requestErrors = new client.Counter({
  name: "vedo_request_errors_total",
  help: "Total failed requests",
  labelNames: ["method", "path"]
})
const requestDuration = new client.Histogram({
  name: "vedo_request_duration_seconds",
  help: "Request duration",
  labelNames: ["method", "path"]
})
register.registerMetric(requestTotal)
register.registerMetric(requestErrors)
register.registerMetric(requestDuration)

const redact = (input: string): string => input.replace(/password|token|secret/gi, "[REDACTED]")

const app = express()

app.use((req: Request, res: Response, next: NextFunction) => {
  const start = Date.now()
  const correlationId = req.header("x-correlation-id") ?? randomUUID()
  res.setHeader("x-correlation-id", correlationId)
  res.on("finish", () => {
    const durationSeconds = (Date.now() - start) / 1000
    requestTotal.inc({ method: req.method, path: req.path, status: String(res.statusCode) })
    requestDuration.observe({ method: req.method, path: req.path }, durationSeconds)
    if (res.statusCode >= 400) {
      requestErrors.inc({ method: req.method, path: req.path })
    }
    logger.info({
      message: redact("request_complete"),
      method: req.method,
      path: req.path,
      status: res.statusCode,
      trace_id: req.header("traceparent") ?? "00000000000000000000000000000000",
      correlation_id: correlationId
    })
  })
  next()
})

app.get("/", (_req: Request, res: Response) => {
  res.json({ name: "service-ts-template", version: "0.1.0", description: "TypeScript service template with observability", stub: true })
})

app.get("/health", (_req: Request, res: Response) => {
  res.json({ status: "healthy" })
})

app.get("/ready", (_req: Request, res: Response) => {
  res.json({ status: "ready" })
})

app.get("/metrics", async (_req: Request, res: Response) => {
  res.set("Content-Type", register.contentType)
  res.send(await register.metrics())
})

app.listen(8080, "0.0.0.0", () => {
  logger.info({ message: "service-ts-template started", port: 8080 })
})
