import http from 'k6/http';
import { check, sleep } from 'k6';

export let options = {
    stages: [
        { duration: '5m', target: 200 },  // Ramp-up to 200 VUs in 5 minutes
        { duration: '10m', target: 1000 }, // Ramp-up to 1000 VUs in 10 minutes
        { duration: '10m', target: 1000 }, // Maintain 1000 VUs for 10 minutes (max load)
        { duration: '5m', target: 2000 }, // Spike to 2000 VUs in 5 minutes
        { duration: '2m', target: 2000 }, // Hold at 2000 VUs for 2 minutes (traffic spike)
        { duration: '5m', target: 1000 }, // Scale down to 1000 VUs in 5 minutes
        { duration: '5m', target: 0 },    // Gradually ramp-down to 0 VUs
    ],
};

const API_URL = 'http://13.233.233.119';

export default function () {
    // Home
    let readRes = http.get(`${API_URL}/`);
    check(readRes, { 'read status was 200': (r) => r.status === 200 });

    // Create
    let createPayload = JSON.stringify({ id: '1', value: 'test value' });
    let createHeaders = { 'Content-Type': 'application/json' };
    let createRes = http.post(`${API_URL}/create`, createPayload, { headers: createHeaders });
    check(createRes, { 'create status was 201': (r) => r.status === 201 });

    // Update
    let updatePayload = JSON.stringify({ id: '1', value: 'updated value' });
    let updateHeaders = { 'Content-Type': 'application/json' };
    let updateRes = http.put(`${API_URL}/update`, updatePayload, { headers: updateHeaders });
    check(updateRes, { 'update status was 200': (r) => r.status === 200 });

    // Delete
    let deleteRes = http.del(`${API_URL}/delete?id=1`);
    check(deleteRes, { 'delete status was 200': (r) => r.status === 200 });

    // Pause between iterations to simulate real user activity
    sleep(1);
}
