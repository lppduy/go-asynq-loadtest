import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

const errorRate = new Rate('errors');

// Stress test - gradually increase load to find breaking point
export const options = {
  stages: [
    { duration: '1m', target: 50 },    // Ramp to 50 users
    { duration: '2m', target: 100 },   // Ramp to 100 users
    { duration: '2m', target: 200 },   // Ramp to 200 users
    { duration: '2m', target: 300 },   // Ramp to 300 users - stress point
    { duration: '2m', target: 400 },   // Push to 400 users - breaking point?
    { duration: '1m', target: 0 },     // Ramp down
  ],
  thresholds: {
    'http_req_duration': ['p(95)<1000'], // 95% under 1s (more lenient)
    'http_req_failed': ['rate<0.1'],     // Error rate under 10%
  },
};

const BASE_URL = 'http://localhost:8080';

function generateOrder(userId) {
  return {
    customer_id: `stress-${userId}`,
    customer_email: `stress${userId}@test.com`,
    items: [
      {
        product_id: `prod-${Math.floor(Math.random() * 100)}`,
        product_name: `Product ${Math.floor(Math.random() * 100)}`,
        quantity: Math.floor(Math.random() * 5) + 1,
        unit_price: Math.floor(Math.random() * 1000) + 10,
      }
    ],
    shipping_address: {
      street: `${Math.floor(Math.random() * 999)} Stress St`,
      city: 'San Francisco',
      state: 'CA',
      postal_code: '94102',
      country: 'USA',
    },
    payment_method: 'credit_card',
    notes: 'Stress test order',
  };
}

export default function () {
  const userId = Math.floor(Math.random() * 100000);
  
  // Create order
  const orderPayload = JSON.stringify(generateOrder(userId));
  const orderRes = http.post(
    `${BASE_URL}/api/v1/orders`,
    orderPayload,
    {
      headers: { 'Content-Type': 'application/json' },
      timeout: '10s',
    }
  );

  const success = check(orderRes, {
    'order created': (r) => r.status === 201,
    'response time acceptable': (r) => r.timings.duration < 2000,
  });

  if (!success) {
    errorRate.add(1);
  }

  sleep(0.5);
}

export function handleSummary(data) {
  return {
    'stdout': textSummary(data),
    'loadtest/results/stress-test-summary.json': JSON.stringify(data),
  };
}

function textSummary(data) {
  let summary = '\n';
  summary += `✅ Stress Test Summary\n`;
  summary += `${'='.repeat(50)}\n\n`;
  
  summary += `📊 Total Requests: ${data.metrics.http_reqs.values.count}\n`;
  summary += `⚡ Request Rate: ${data.metrics.http_reqs.values.rate.toFixed(2)} req/s\n`;
  summary += `⏱️  Avg Response: ${data.metrics.http_req_duration.values.avg.toFixed(2)}ms\n`;
  summary += `🔥 Max Response: ${data.metrics.http_req_duration.values.max.toFixed(2)}ms\n`;
  summary += `❌ Failed: ${data.metrics.http_req_failed.values.passes || 0}\n`;
  summary += `👥 Max Users: ${data.metrics.vus_max.values.max}\n\n`;
  
  return summary;
}
