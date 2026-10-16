import { DurableObject } from "cloudflare:workers";

interface Env {
  GRAPHFOLIO_CONTAINER: DurableObjectNamespace<GraphFolioContainer>;
  PORTFOLIO_HYPERDRIVE?: Hyperdrive;
  USER_HYPERDRIVE?: Hyperdrive;
  CORS_ORIGINS: string;
  INTERNAL_CRON_SECRET?: string;
}

// -----------------------------------------------------------------------------
// Durable Object housing the Cloudflare Container instance
// -----------------------------------------------------------------------------
export class GraphFolioContainer extends DurableObject {
  async fetch(request: Request): Promise<Response> {
    const url = new URL(request.url);

    // Forward the request to the container's BFF HTTP port (:8080)
    url.protocol = "http:";
    url.host = "127.0.0.1:8080";

    const forwardedRequest = new Request(url.toString(), {
      method: request.method,
      headers: request.headers,
      body: request.body,
      redirect: "manual",
    });

    // Strip incoming host header so container matches on local port
    forwardedRequest.headers.delete("host");

    try {
      const tcpPort = (this.ctx as any).container.getTcpPort(8080);
      return await tcpPort.fetch(forwardedRequest);
    } catch (err: any) {
      return new Response(
        JSON.stringify({
          error: "Container upstream unavailable",
          details: err?.message || String(err),
        }),
        {
          status: 503,
          headers: { "Content-Type": "application/json" },
        }
      );
    }
  }
}

// -----------------------------------------------------------------------------
// Worker Gateway Router
// -----------------------------------------------------------------------------
export default {
  async fetch(request: Request, env: Env, ctx: ExecutionContext): Promise<Response> {
    const url = new URL(request.url);
    const origin = request.headers.get("Origin") || "";
    const allowedOrigins = (env.CORS_ORIGINS || "").split(",").map((o) => o.trim());
    const isAllowedOrigin = allowedOrigins.includes(origin);

    // 1. Handle CORS Preflight at Edge
    if (request.method === "OPTIONS") {
      if (isAllowedOrigin) {
        return new Response(null, {
          status: 204,
          headers: {
            "Access-Control-Allow-Origin": origin,
            "Access-Control-Allow-Methods": "GET, POST, OPTIONS, PUT, DELETE",
            "Access-Control-Allow-Headers": "Content-Type, Authorization, X-Requested-With, X-CSRF-Token",
            "Access-Control-Allow-Credentials": "true",
            "Access-Control-Max-Age": "86400",
          },
        });
      }
      return new Response(null, { status: 204 });
    }

    // 2. Health check probe for edge diagnostics
    if (url.pathname === "/healthz") {
      return new Response(
        JSON.stringify({ status: "healthy", service: "graphfolio-edge-gateway", timestamp: new Date().toISOString() }),
        {
          status: 200,
          headers: { "Content-Type": "application/json" },
        }
      );
    }

    // 3. Route to GraphFolio Container Instance
    const id = env.GRAPHFOLIO_CONTAINER.idFromName("graphfolio-production-instance");
    const containerStub = env.GRAPHFOLIO_CONTAINER.get(id);

    const response = await containerStub.fetch(request);

    // 4. Inject CORS headers on outgoing response
    const modifiedResponse = new Response(response.body, response);
    if (isAllowedOrigin) {
      modifiedResponse.headers.set("Access-Control-Allow-Origin", origin);
      modifiedResponse.headers.set("Access-Control-Allow-Credentials", "true");
    }

    return modifiedResponse;
  },

  // ---------------------------------------------------------------------------
  // Scheduled Cron Handler for EOD Valuation Job
  // ---------------------------------------------------------------------------
  async scheduled(event: ScheduledEvent, env: Env, ctx: ExecutionContext): Promise<void> {
    console.log(`[cron] Triggering Daily EOD Valuation Job at ${new Date(event.scheduledTime).toISOString()}`);

    const id = env.GRAPHFOLIO_CONTAINER.idFromName("graphfolio-production-instance");
    const containerStub = env.GRAPHFOLIO_CONTAINER.get(id);

    const internalJobReq = new Request("http://127.0.0.1:8080/internal/jobs/valuation-eod", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "X-Internal-Secret": env.INTERNAL_CRON_SECRET || "",
      },
      body: JSON.stringify({
        asOf: new Date(event.scheduledTime).toISOString().slice(0, 10),
      }),
    });

    try {
      const res = await containerStub.fetch(internalJobReq);
      if (!res.ok) {
        console.error(`[cron] Valuation job failed with status ${res.status}: ${await res.text()}`);
      } else {
        console.log(`[cron] Valuation job completed successfully: ${await res.text()}`);
      }
    } catch (err: any) {
      console.error(`[cron] Failed to dispatch valuation job to container: ${err?.message || err}`);
    }
  },
};
