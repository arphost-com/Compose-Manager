export function formFromDockerConfig(config) {
  const logOpts = config['log-opts'] || {};
  const hosts = Array.isArray(config.hosts) ? config.hosts : [];
  const tcpHost = hosts.find(host => typeof host === 'string' && host.startsWith('tcp://')) || '';
  const tcpMatch = tcpHost.match(/^tcp:\/\/([^:]+):(\d+)$/);
  return {
    live_restore: Boolean(config['live-restore']),
    log_driver: String(config['log-driver'] || 'json-file'),
    log_max_size: String(logOpts['max-size'] || '10m'),
    log_max_file: String(logOpts['max-file'] || '3'),
    dns: arrayToLines(config.dns),
    registry_mirrors: arrayToLines(config['registry-mirrors']),
    insecure_registries: arrayToLines(config['insecure-registries']),
    default_address_pools: poolsToLines(config['default-address-pools']),
    bip: String(config.bip || ''),
    ipv6: Boolean(config.ipv6),
    fixed_cidr_v6: String(config['fixed-cidr-v6'] || ''),
    expose_tcp: Boolean(tcpHost),
    tcp_bind: tcpMatch ? tcpMatch[1] : '127.0.0.1',
    tcp_port: tcpMatch ? tcpMatch[2] : '2376',
  };
}

export function buildDockerConfig(raw, form) {
  let config = {};
  if (String(raw || '').trim()) {
    config = JSON.parse(raw);
  }
  if (!config || Array.isArray(config) || typeof config !== 'object') {
    throw new Error('daemon.json must be a JSON object');
  }
  config['live-restore'] = Boolean(form.live_restore);
  config.ipv6 = Boolean(form.ipv6);
  setOrDelete(config, 'fixed-cidr-v6', form.fixed_cidr_v6);
  setOrDelete(config, 'dns', linesToArray(form.dns));
  setOrDelete(config, 'registry-mirrors', linesToArray(form.registry_mirrors));
  setOrDelete(config, 'insecure-registries', linesToArray(form.insecure_registries));
  setOrDelete(config, 'default-address-pools', linesToPools(form.default_address_pools));
  setOrDelete(config, 'bip', String(form.bip || '').trim());

  if (form.log_driver) {
    config['log-driver'] = form.log_driver;
    config['log-opts'] = {
      ...(config['log-opts'] || {}),
      'max-size': form.log_max_size || '10m',
      'max-file': form.log_max_file || '3',
    };
  }

  const hosts = Array.isArray(config.hosts) ? config.hosts.filter(host => typeof host !== 'string' || !host.startsWith('tcp://')) : [];
  if (form.expose_tcp) {
    if (!hosts.some(host => typeof host === 'string' && (host.startsWith('unix://') || host === 'fd://'))) {
      hosts.unshift('unix:///var/run/docker.sock');
    }
    hosts.push(`tcp://${form.tcp_bind || '127.0.0.1'}:${form.tcp_port || '2376'}`);
  }
  setOrDelete(config, 'hosts', hosts);
  return config;
}

export function pruneMap(value) {
  return Object.fromEntries(Object.entries(value || {}).filter(([, v]) => String(v || '').trim() !== ''));
}

function setOrDelete(object, key, value) {
  if (Array.isArray(value) && value.length === 0) {
    delete object[key];
    return;
  }
  if (typeof value === 'string' && value.trim() === '') {
    delete object[key];
    return;
  }
  object[key] = value;
}

function arrayToLines(value) {
  return Array.isArray(value) ? value.join('\n') : '';
}

function linesToArray(value) {
  return String(value || '').split(/\r?\n/).map(item => item.trim()).filter(Boolean);
}

function poolsToLines(value) {
  if (!Array.isArray(value)) return '';
  return value.map(pool => `${pool.base || ''},${pool.size || ''}`).filter(line => line !== ',').join('\n');
}

function linesToPools(value) {
  const lines = linesToArray(value);
  return lines.map((line, index) => {
    const parts = line.split(',').map(part => part.trim());
    const base = parts[0] || '';
    let sizeText = parts[1] || '';
    if (!base) throw new Error(`Default address pool line ${index + 1} is empty.`);
    if (!/^\d+\.\d+\.\d+\.\d+\/\d+$/.test(base)) {
      throw new Error(`Default address pool line ${index + 1}: enter a CIDR like 172.30.0.0/16 (got "${base}").`);
    }
    const baseMask = Number(base.split('/')[1]);
    // Single-CIDR form (no comma): treat the CIDR itself as a one-subnet pool.
    // Advanced form (base,size): base is the outer pool and size is the subnet
    // size Docker slices out. 172.30.0.0/16,24 gives 256 /24 subnets from a /16.
    if (!sizeText) {
      sizeText = String(baseMask);
    }
    const size = Number(sizeText);
    if (!Number.isInteger(size) || size < 1 || size > 32) {
      throw new Error(`Default address pool line ${index + 1}: size must be an integer 1-32 (got "${sizeText}").`);
    }
    if (size < baseMask) {
      throw new Error(`Default address pool line ${index + 1}: subnet size /${size} must be greater than or equal to the base mask /${baseMask}.`);
    }
    return { base, size };
  });
}
