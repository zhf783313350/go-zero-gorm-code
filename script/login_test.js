import http from 'k6/http';
import { check, sleep } from 'k6';

// 1. 压测梯度与考核指标配置
export const options = {
  stages: [
    { duration: '10s', target: 20 },  // 10秒内逐渐拉升到 20 个并发用户
    { duration: '30s', target: 100 }, // 保持/攀升至 100 个并发运行 30 秒
    { duration: '10s', target: 0 },   // 10秒内降至 0，完成平滑回落
  ],
  thresholds: {
    http_req_duration: ['p(95)<2000'],
    http_req_failed: ['rate<0.01'],
  },
};

// 2. 核心登录请求逻辑
export default function () {
  // 根据你捕获到的网络请求信息配置 URL
  const url = 'http://localhost:8080/api/user/login';

  // 对应你截图和日志中的请求 Payload
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

  // 每个虚拟用户请求后停顿 1 秒，模拟真实用户交互间隙
  sleep(1);
}