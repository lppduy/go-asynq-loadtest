import http from 'k6/http';
import { check, sleep } from 'k6';

// Spike test - sudden traffic spike to test recovery
export const options = {
  stages: [
    { duration: '30s', target: 10 },   // Warm up
    { duration: '10s', target: 200 },  // SPIKE! 10 → 200 users trong 10s
    { duration: '1m', target: 200 },   // Stay at spike
    { duration: '10s', target: 10 },   // Drop back down
    { duration: '30s', target: 0 },    // Cool down
  ],
  thresholds: {
    'http_req_duration': ['p(95)<2000'], // More lenient threshold for spike
    'http_req_failed': ['rate<0.2'],     // Allow up to 20% errors during spike
  },
};

const BASE_URL = 'http://localhost:8080';

function generateOrder(userId) {
  return {
    customer_id: `spike-${userId}`,
    customer_email: `spike${userId}@test.com`,
    items: [{
      product_id: `prod-${Math.floor(Math.random() * 50)}`,
      product_name: `Spike Product`,
      quantity: 1,
      unit_price: 99.99,
    }],
    shipping_address: {
      street: '999 Spike St',
      city: 'SF',
      state: 'CA',
      postal_code: '94102',
      country: 'USA',
    },
    payment_method: 'credit_card',
    notes: 'Spike test - sudden load',
  };
}

export default function () {
  const userId = Math.floor(Math.random() * 50000);
  
  const orderPayload = JSON.stringify(generateOrder(userId));
  const orderRes = http.post(
    `${BASE_URL}/api/v1/orders`,
    orderPayload,
    {
      headers: { 'Content-Type': 'application/json' },
      timeout: '15s',
    }
  );

  check(orderRes, {
    'order created': (r) => r.status === 201,
    'responded in time': (r) => r.timings.duration < 5000,
  });

  sleep(0.3);
}

export function handleSummary(data) {
  return {
    'stdout': textSummary(data),
    'loadtest/results/spike-test-summary.json': JSON.stringify(data),
  };
}

function textSummary(data) {
  let summary = '\n';
  summary += `⚡ Spike Test Summary\n`;
  summary += `${'='.repeat(50)}\n\n`;
  
  summary += `📊 Total Requests: ${data.metrics.http_reqs.values.count}\n`;
  summary += `⚡ Peak Request Rate: ${data.metrics.http_reqs.values.rate.toFixed(2)} req/s\n`;
  summary += `⏱️  Avg Response: ${data.metrics.http_req_duration.values.avg.toFixed(2)}ms\n`;
  summary += `🔥 p(95) Response: ${data.metrics.http_req_duration.values['p(95)'].toFixed(2)}ms\n`;
  summary += `❌ Failed: ${data.metrics.http_req_failed.values.passes || 0} (${((data.metrics.http_req_failed.values.rate || 0) * 100).toFixed(2)}%)\n`;
  summary += `👥 Max Concurrent Users: ${data.metrics.vus_max.values.max}\n\n`;
  
  summary += `🎯 Key Observations:\n`;
  summary += `  - System handled sudden ${data.metrics.vus_max.values.max}x spike\n`;
  summary += `  - Recovery time: Check metrics during ramp down\n`;
  summary += `  - Queue depth: Check Asynqmon during spike\n\n`;
  
  return summary;
}
