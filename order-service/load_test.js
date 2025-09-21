import http from 'k6/http';
import { check } from 'k6';

export const options = {
    scenarios: {
        constant_arrival_rate: {
            executor: 'constant-arrival-rate',
            rate: 1000,
            duration: '30s',
            preAllocatedVUs: 1000,
            maxVUs: 2000,
        },
    },
};

export default function () {
    const payload = JSON.stringify({
        product_id: 1,
        qty: 1,
    });

    const headers = {
        'Content-Type': 'application/json',
    };

    const res = http.post('http://order-service:8080/orders', payload, { headers: headers });

    check(res, {
        'status is 202': (r) => r.status === 202
    });
}