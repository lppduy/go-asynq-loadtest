import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');

// Test configuration
export const options = {
  stages: [
    { duration: '30s', target: 20 },  // Ramp up: 0 → 20 users trong 30s
    { duration: '1m', target: 50 },   // Ramp up: 20 → 50 users trong 1m
    { duration: '2m', target: 50 },   // Stay: 50 users trong 2m (peak load)
    { duration: '30s', target: 0 },   // Ramp down: 50 → 0 users trong 30s
  ],
  thresholds: {
    'http_req_duration': ['p(95)<500'], // 95% requests phải < 500ms
    'http_req_failed': ['rate<0.05'],   // Error rate < 5%
    'errors': ['rate<0.1'],             // Custom error < 10%
  },
};

const BASE_URL = 'http://localhost:8080';

// Generate random order data
function generateOrder(userId) {
  return {
    customer_id: `load-test-${userId}`,
    customer_email: `user${userId}@loadtest.com`,
    items: [
      {
        product_id: `prod-${Math.floor(Math.random() * 100)}`,
        product_name: `Product ${Math.floor(Math.random() * 100)}`,
        quantity: Math.floor(Math.random() * 5) + 1,
        unit_price: Math.floor(Math.random() * 1000) + 10,
      }
    ],
    shipping_address: {
      street: `${Math.floor(Math.random() * 999)} Main St`,
      city: 'San Francisco',
      state: 'CA',
      postal_code: '94102',
      country: 'USA',
    },
    payment_method: 'credit_card',
    notes: `Load test order created at ${new Date().toISOString()}`,
  };
}

// Main test function - K6 sẽ chạy function này cho mỗi virtual user
export default function () {
  const userId = Math.floor(Math.random() * 10000);
  
  // Test 1: Health check
  const healthRes = http.get(`${BASE_URL}/health`);
  check(healthRes, {
    'health check status is 200': (r) => r.status === 200,
  }) || errorRate.add(1);

  sleep(0.5);

  // Test 2: Create order (main test)
  const orderPayload = JSON.stringify(generateOrder(userId));
  const orderRes = http.post(
    `${BASE_URL}/api/v1/orders`,
    orderPayload,
    {
      headers: { 'Content-Type': 'application/json' },
    }
  );

  const orderCreated = check(orderRes, {
    'order created status is 201': (r) => r.status === 201,
    'order has ID': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.id && body.id.startsWith('ORD-');
      } catch (e) {
        return false;
      }
    },
    'response time < 200ms': (r) => r.timings.duration < 200,
  });

  if (!orderCreated) {
    errorRate.add(1);
    console.error(`Failed to create order: ${orderRes.status} - ${orderRes.body}`);
  } else {
    // Test 3: Get order by ID
    try {
      const order = JSON.parse(orderRes.body);
      const getRes = http.get(`${BASE_URL}/api/v1/orders/${order.id}`);
      
      check(getRes, {
        'get order status is 200': (r) => r.status === 200,
        'order data matches': (r) => {
          try {
            const retrieved = JSON.parse(r.body);
            return retrieved.id === order.id;
          } catch (e) {
            return false;
          }
        },
      }) || errorRate.add(1);
    } catch (e) {
      console.error(`Error parsing order response: ${e}`);
      errorRate.add(1);
    }
  }

  sleep(1);

  // Test 4: List orders
  const listRes = http.get(`${BASE_URL}/api/v1/orders?page=1&limit=10`);
  check(listRes, {
    'list orders status is 200': (r) => r.status === 200,
    'list has orders': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.total > 0;
      } catch (e) {
        return false;
      }
    },
  }) || errorRate.add(1);

  sleep(0.5);
}

// Summary function (hiển thị khi test xong)
export function handleSummary(data) {
  return {
    'stdout': textSummary(data, { indent: ' ', enableColors: true }),
    'loadtest/results/basic-load-summary.json': JSON.stringify(data),
  };
}

function textSummary(data, options = {}) {
  const indent = options.indent || '';
  
  let summary = '\n';
  summary += `${indent}✅ Basic Load Test Summary\n`;
  summary += `${indent}${'='.repeat(50)}\n\n`;
  
  summary += `${indent}📊 Requests:\n`;
  summary += `${indent}  Total: ${data.metrics.http_reqs.values.count}\n`;
  summary += `${indent}  Rate: ${data.metrics.http_reqs.values.rate.toFixed(2)} req/s\n\n`;
  
  summary += `${indent}⏱️  Response Time:\n`;
  summary += `${indent}  Avg: ${data.metrics.http_req_duration.values.avg.toFixed(2)}ms\n`;
  summary += `${indent}  Min: ${data.metrics.http_req_duration.values.min.toFixed(2)}ms\n`;
  summary += `${indent}  Max: ${data.metrics.http_req_duration.values.max.toFixed(2)}ms\n`;
  summary += `${indent}  p(95): ${data.metrics.http_req_duration.values['p(95)'].toFixed(2)}ms\n`;
  summary += `${indent}  p(99): ${data.metrics.http_req_duration.values['p(99)'].toFixed(2)}ms\n\n`;
  
  summary += `${indent}❌ Errors:\n`;
  summary += `${indent}  Failed Requests: ${data.metrics.http_req_failed.values.passes || 0}\n`;
  summary += `${indent}  Error Rate: ${((data.metrics.errors?.values.rate || 0) * 100).toFixed(2)}%\n\n`;
  
  summary += `${indent}👥 Virtual Users:\n`;
  summary += `${indent}  Max: ${data.metrics.vus_max.values.max}\n\n`;
  
  return summary;
}
