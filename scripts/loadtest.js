import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
    stages: [
        { duration: '1m', target: 50 },   // разогрев
        { duration: '2m', target: 100 },  // пик
        { duration: '1m', target: 0 },    // спад
    ],
    thresholds: {
        http_req_duration: ['p(95)<500'], // 95% запросов <500ms
        http_req_failed: ['rate<0.01'],
    },
};

const BASE_URL = 'http://localhost:8080/api';

export function setup() {
    const loginRes = http.post(`${BASE_URL}/login`, JSON.stringify({
        login: 'user1@example.com',
        password: '12345678',
    }), { headers: { 'Content-Type': 'application/json' } });
    check(loginRes, { 'login success': (r) => r.status === 200 });
    const cookie = loginRes.headers['Set-Cookie'];
    return { cookie };
}

export default function (data) {
    const headers = { Cookie: data.cookie, 'Content-Type': 'application/json' };
    
    // 1. Список мест с фильтрацией
    let res = http.get(`${BASE_URL}/places?category_ids=1,2&min_rating=4.0&min_reviews=10`, { headers });
    check(res, { 'places list status 200': (r) => r.status === 200 });
    sleep(1);
    
    // 2. Детали места
    let placeId = Math.floor(Math.random() * 100000) + 1;
    res = http.get(`${BASE_URL}/places/${placeId}`, { headers });
    check(res, { 'place details status 200': (r) => r.status === 200 });
    sleep(1);
    
    // 3. Поиск мест
    res = http.get(`${BASE_URL}/places/search?q=eiffel`, { headers });
    check(res, { 'search status 200': (r) => r.status === 200 });
    sleep(1);
    
    // 4. Список поездок
    res = http.get(`${BASE_URL}/trips`, { headers });
    check(res, { 'trips list status 200': (r) => r.status === 200 });
    sleep(1);
    
    // 5. Детали поездки
    let tripId = Math.floor(Math.random() * 30000) + 1;
    res = http.get(`${BASE_URL}/trips/${tripId}`, { headers });
    check(res, { 'trip details status 200': (r) => r.status === 200 });
    sleep(1);
}