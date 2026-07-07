import test from 'node:test';
import assert from 'node:assert/strict';
import { buildDockerConfig, formFromDockerConfig, pruneMap } from './dockerSettings.js';

const baseForm = {
  live_restore: false,
  log_driver: 'json-file',
  log_max_size: '10m',
  log_max_file: '3',
  dns: '',
  registry_mirrors: '',
  insecure_registries: '',
  default_address_pools: '',
  bip: '',
  ipv6: false,
  fixed_cidr_v6: '',
  expose_tcp: false,
  tcp_bind: '127.0.0.1',
  tcp_port: '2376',
};

test('buildDockerConfig preserves valid default-address-pools entries', () => {
  const config = buildDockerConfig('{}', {
    ...baseForm,
    default_address_pools: '172.30.0.0/16,24\n10.88.0.0/16,24',
  });

  assert.deepEqual(config['default-address-pools'], [
    { base: '172.30.0.0/16', size: 24 },
    { base: '10.88.0.0/16', size: 24 },
  ]);
});

test('buildDockerConfig accepts single-CIDR default-address-pools entries', () => {
  const config = buildDockerConfig('{}', {
    ...baseForm,
    default_address_pools: '172.30.0.0/16',
  });

  assert.deepEqual(config['default-address-pools'], [
    { base: '172.30.0.0/16', size: 16 },
  ]);
});

test('buildDockerConfig rejects malformed default-address-pools per line', () => {
  assert.throws(
    () => buildDockerConfig('{}', {
      ...baseForm,
      default_address_pools: '172.30.0.0/16,24\nnot-a-cidr,24',
    }),
    /Default address pool line 2: enter a CIDR/
  );
});

test('buildDockerConfig rejects subnet sizes smaller than the base mask', () => {
  assert.throws(
    () => buildDockerConfig('{}', {
      ...baseForm,
      default_address_pools: '172.30.0.0/16,12',
    }),
    /subnet size \/12 must be greater than or equal to the base mask \/16/
  );
});

test('formFromDockerConfig round-trips default-address-pools', () => {
  const form = formFromDockerConfig({
    'default-address-pools': [
      { base: '172.30.0.0/16', size: 24 },
      { base: '10.88.0.0/16', size: 24 },
    ],
  });

  assert.equal(form.default_address_pools, '172.30.0.0/16,24\n10.88.0.0/16,24');
});

test('pruneMap drops blank values but keeps populated fields', () => {
  assert.deepEqual(pruneMap({ host: 'backup.example.com', port: '', username: 'alice' }), {
    host: 'backup.example.com',
    username: 'alice',
  });
});
