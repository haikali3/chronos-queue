
import grpc from 'k6/net/grpc';
import { check } from 'k6';

const client = new grpc.Client();
client.load(['../../proto'], 'producer.proto');

export const options = {
  stages: [
    { duration: '30s', target: 10 }, // Ramp up to 10 users over 30 seconds
    { duration: '1m', target: 10 },  // Stay at 10 users for 1 minute
    { duration: '30s', target: 0 },   // Ramp down to 0 users over 30 seconds
  ],
}
export default function () {
  
  // connect invoke check close
  client.connect('localhost:50051', {
    plaintext: true,
  });

  const response = client.invoke('chronos.ProducerService/SubmitJob', {
    type: 'load-k6-test',
    payload: 'dGVzdA==',  // base64 for bytes
    max_retries: 3,
    idempotencyKey: `k6-test-${__VU}-${__ITER}`,  // Unique key per VU and iteration
  });

  check(response, {
    'status is OK': (r) => r && r.status === grpc.StatusOK,
  });

  client.close();

}