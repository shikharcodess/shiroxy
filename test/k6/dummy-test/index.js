import http from 'k6/http';
import { check, sleep } from 'k6';

export let options = {
    vus: 1000, // number of virtual users
    duration: '10s', // test duration
};

const API_URL = 'http://127.0.0.1:80';

export default function () {
    // Home
    let readRes = http.get(`${API_URL}/`);
    check(readRes, { 'read status was 200': (r) => r.status === 200 });

    // // Create
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
