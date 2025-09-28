-- policies (можно хранить YAML целиком + индекс по tenant/version)
create table policies (
  tenant_id text not null,
  version text not null,
  yaml text not null,
  status text not null check (status in ('draft','active','deprecated')),
  created_at timestamptz not null default now(),
  activated_at timestamptz,
  primary key (tenant_id, version)
);

-- moderation_requests
create table moderation_requests (
  id uuid primary key,
  tenant_id text not null,
  request_id text,
  policy_version text not null,
  content_sha256 bytea not null,
  content_preview text,
  lang text,
  status text not null check (status in ('pending','done','escalated','failed')),
  created_at timestamptz not null default now(),
  completed_at timestamptz
);

-- moderation_verdicts
create table moderation_verdicts (
  req_id uuid primary key references moderation_requests(id) on delete cascade,
  action text not null check (action in ('allow','soft_block','block','escalate')),
  severity text,
  categories jsonb not null,
  explain jsonb not null
);

create index on policies(tenant_id, status);
create index idx_req_tenant_created on moderation_requests(tenant_id, created_at desc);
