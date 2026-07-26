import http from 'k6/http';
import { check } from 'k6';

// 1. 精确控制：目标一秒钟 500 次请求 (500 RPS)
export const options = {
  scenarios: {
    contacts: {
      executor: 'ramping-arrival-rate', // 使用速率控制执行器
      startRate: 50,                    // 初始每秒 50 请求
      timeUnit: '1s',                   // 时间单位：1 秒
      preAllocatedVUs: 100,             // 预先分配 100 个虚拟用户
      maxVUs: 1000,                     // 如果接口变慢，最多自动扩容到 1000 用户来维持 500 RPS
      stages: [
        { duration: '10s', target: 100 }, // 10秒内提升到 每秒 500 次请求
        { duration: '30s', target: 100 }, // 稳定保持 每秒 500 次请求 运行 30秒
        { duration: '10s', target: 0 },   // 10秒内降回 0
      ],
    },
  },
  thresholds: {
    http_req_duration: ['p(95)<1000'],  // 考核：95% 的请求耗时在 1 秒内
    http_req_failed: ['rate<0.01'],     // 考核：失败率小于 1%
  },
};

// 2. 核心登录请求逻辑
export default function () {
  const url = 'http://localhost:8080/api/user/login';

  const payload = JSON.stringify({
    phoneNumber: '13800000001',
    password: '123456',
  });

  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
  };

  // 发起 POST 登录请求
  const res = http.post(url, payload, params);

  // 断言：检查接口返回的状态码是否为 200 OK
  check(res, {
    'status is 200': (r) => r.status === 200,
  });

  // 注意：此处不写 sleep()，k6 会自动精准接管并强制维持一秒钟 500 次的频率！
}