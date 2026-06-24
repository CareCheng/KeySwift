CREATE TABLE IF NOT EXISTS schema_metadata (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    schema_key TEXT NOT NULL UNIQUE,
    schema_version TEXT NOT NULL,
    schema_checksum TEXT NOT NULL,
    app_version TEXT NOT NULL,
    initialized_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS schema_revisions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    schema_key TEXT NOT NULL,
    version TEXT NOT NULL,
    name TEXT NOT NULL DEFAULT '',
    direction TEXT NOT NULL DEFAULT 'baseline',
    checksum TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'applied',
    app_version TEXT NOT NULL DEFAULT '',
    started_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    finished_at DATETIME,
    error_message TEXT NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(schema_key, version, direction)
);

CREATE INDEX IF NOT EXISTS idx_schema_revisions_schema_key ON schema_revisions(schema_key);
CREATE INDEX IF NOT EXISTS idx_schema_revisions_status ON schema_revisions(status);

CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL DEFAULT '',
    phone TEXT NOT NULL DEFAULT '',
    email_verified INTEGER NOT NULL DEFAULT 0,
    enable2_fa INTEGER NOT NULL DEFAULT 0,
    totp_secret TEXT NOT NULL DEFAULT '',
    prefer_email_auth INTEGER NOT NULL DEFAULT 1,
    pay_password TEXT NOT NULL DEFAULT '',
    pay_password_set INTEGER NOT NULL DEFAULT 0,
    pay_password_errors INTEGER NOT NULL DEFAULT 0,
    pay_password_lock_at DATETIME,
    status INTEGER NOT NULL DEFAULT 1,
    last_login_at DATETIME,
    last_login_ip TEXT NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME
);

CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at);

CREATE TABLE IF NOT EXISTS email_verify_codes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT NOT NULL,
    code TEXT NOT NULL,
    type TEXT NOT NULL,
    used INTEGER NOT NULL DEFAULT 0,
    expires_at DATETIME NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_email_verify_codes_email ON email_verify_codes(email);

CREATE TABLE IF NOT EXISTS product_categories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL DEFAULT '',
    icon TEXT NOT NULL DEFAULT '',
    sort_order INTEGER NOT NULL DEFAULT 0,
    status INTEGER NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME
);

CREATE INDEX IF NOT EXISTS idx_product_categories_deleted_at ON product_categories(deleted_at);

CREATE TABLE IF NOT EXISTS products (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    detail TEXT NOT NULL DEFAULT '',
    specs TEXT NOT NULL DEFAULT '',
    features TEXT NOT NULL DEFAULT '',
    tags TEXT NOT NULL DEFAULT '',
    price REAL NOT NULL DEFAULT 0,
    duration INTEGER NOT NULL DEFAULT 0,
    duration_unit TEXT NOT NULL DEFAULT '天',
    stock INTEGER NOT NULL DEFAULT 0,
    status INTEGER NOT NULL DEFAULT 1,
    sort_order INTEGER NOT NULL DEFAULT 0,
    category_id INTEGER NOT NULL DEFAULT 0,
    product_type INTEGER NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME
);

CREATE INDEX IF NOT EXISTS idx_products_deleted_at ON products(deleted_at);
CREATE INDEX IF NOT EXISTS idx_products_category_id ON products(category_id);

CREATE TABLE IF NOT EXISTS orders (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    order_no TEXT NOT NULL UNIQUE,
    payment_no TEXT NOT NULL DEFAULT '',
    buyer_subject_id TEXT NOT NULL DEFAULT '',
    user_id INTEGER NOT NULL DEFAULT 0,
    username TEXT NOT NULL DEFAULT '',
    product_id INTEGER NOT NULL DEFAULT 0,
    product_name TEXT NOT NULL DEFAULT '',
    quantity INTEGER NOT NULL DEFAULT 1,
    original_price REAL NOT NULL DEFAULT 0,
    price REAL NOT NULL DEFAULT 0,
    paid_amount REAL NOT NULL DEFAULT 0,
    amount_total_cents INTEGER NOT NULL DEFAULT 0,
    amount_payable_cents INTEGER NOT NULL DEFAULT 0,
    amount_paid_cents INTEGER NOT NULL DEFAULT 0,
    currency TEXT NOT NULL DEFAULT 'CNY',
    duration INTEGER NOT NULL DEFAULT 0,
    duration_unit TEXT NOT NULL DEFAULT '',
    status INTEGER NOT NULL DEFAULT 0,
    order_status TEXT NOT NULL DEFAULT 'pending_payment',
    payment_status TEXT NOT NULL DEFAULT 'unpaid',
    delivery_status TEXT NOT NULL DEFAULT 'unfulfilled',
    payment_method TEXT NOT NULL DEFAULT '',
    payment_time DATETIME,
    kami_code TEXT NOT NULL DEFAULT '',
    idempotency_key TEXT NOT NULL DEFAULT '',
    version INTEGER NOT NULL DEFAULT 1,
    remark TEXT NOT NULL DEFAULT '',
    client_ip TEXT NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME
);

CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id);
CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);
CREATE INDEX IF NOT EXISTS idx_orders_buyer_subject_id ON orders(buyer_subject_id);
CREATE INDEX IF NOT EXISTS idx_orders_order_status ON orders(order_status);
CREATE INDEX IF NOT EXISTS idx_orders_payment_status ON orders(payment_status);
CREATE INDEX IF NOT EXISTS idx_orders_delivery_status ON orders(delivery_status);
CREATE INDEX IF NOT EXISTS idx_orders_idempotency_key ON orders(idempotency_key);
CREATE INDEX IF NOT EXISTS idx_orders_deleted_at ON orders(deleted_at);

CREATE TABLE IF NOT EXISTS order_items (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    order_id INTEGER NOT NULL,
    item_type TEXT NOT NULL DEFAULT 'host_product',
    product_ref TEXT NOT NULL DEFAULT '',
    sku_ref TEXT NOT NULL DEFAULT '',
    quantity INTEGER NOT NULL DEFAULT 1,
    unit_price_cents INTEGER NOT NULL DEFAULT 0,
    item_snapshot_json TEXT NOT NULL DEFAULT '',
    delivery_requirement_json TEXT NOT NULL DEFAULT '',
    owner_plugin_id TEXT NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_order_items_order_id ON order_items(order_id);
CREATE INDEX IF NOT EXISTS idx_order_items_product_ref ON order_items(product_ref);
CREATE INDEX IF NOT EXISTS idx_order_items_owner_plugin_id ON order_items(owner_plugin_id);

CREATE TABLE IF NOT EXISTS order_payment_attempts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    attempt_no TEXT NOT NULL UNIQUE,
    order_id INTEGER NOT NULL,
    order_no TEXT NOT NULL DEFAULT '',
    payment_plugin_id TEXT NOT NULL DEFAULT '',
    payment_channel TEXT NOT NULL DEFAULT '',
    amount_cents INTEGER NOT NULL DEFAULT 0,
    currency TEXT NOT NULL DEFAULT 'CNY',
    status TEXT NOT NULL DEFAULT 'created',
    provider_transaction_id TEXT NOT NULL DEFAULT '',
    idempotency_key TEXT NOT NULL DEFAULT '',
    callback_token TEXT NOT NULL DEFAULT '',
    error_code TEXT NOT NULL DEFAULT '',
    error_message TEXT NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(order_id, idempotency_key)
);

CREATE INDEX IF NOT EXISTS idx_order_payment_attempts_order_id ON order_payment_attempts(order_id);
CREATE INDEX IF NOT EXISTS idx_order_payment_attempts_order_no ON order_payment_attempts(order_no);
CREATE INDEX IF NOT EXISTS idx_order_payment_attempts_plugin ON order_payment_attempts(payment_plugin_id);
CREATE INDEX IF NOT EXISTS idx_order_payment_attempts_status ON order_payment_attempts(status);

CREATE TABLE IF NOT EXISTS order_payment_callbacks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    callback_id TEXT NOT NULL UNIQUE,
    attempt_no TEXT NOT NULL,
    order_no TEXT NOT NULL DEFAULT '',
    provider_transaction_id TEXT NOT NULL DEFAULT '',
    provider_status TEXT NOT NULL DEFAULT '',
    amount_cents INTEGER NOT NULL DEFAULT 0,
    currency TEXT NOT NULL DEFAULT 'CNY',
    verified INTEGER NOT NULL DEFAULT 0,
    idempotency_key TEXT NOT NULL DEFAULT '',
    callback_payload_json TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'received',
    error_message TEXT NOT NULL DEFAULT '',
    received_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_order_payment_callbacks_attempt_no ON order_payment_callbacks(attempt_no);
CREATE INDEX IF NOT EXISTS idx_order_payment_callbacks_order_no ON order_payment_callbacks(order_no);
CREATE INDEX IF NOT EXISTS idx_order_payment_callbacks_status ON order_payment_callbacks(status);

CREATE TABLE IF NOT EXISTS order_deliveries (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    delivery_no TEXT NOT NULL UNIQUE,
    order_id INTEGER NOT NULL,
    order_no TEXT NOT NULL DEFAULT '',
    fulfillment_plugin_id TEXT NOT NULL DEFAULT '',
    delivery_type TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'created',
    result_payload_json TEXT NOT NULL DEFAULT '',
    idempotency_key TEXT NOT NULL DEFAULT '',
    delivered_at DATETIME,
    error_code TEXT NOT NULL DEFAULT '',
    error_message TEXT NOT NULL DEFAULT '',
    retryable INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(order_id, idempotency_key)
);

CREATE INDEX IF NOT EXISTS idx_order_deliveries_order_id ON order_deliveries(order_id);
CREATE INDEX IF NOT EXISTS idx_order_deliveries_order_no ON order_deliveries(order_no);
CREATE INDEX IF NOT EXISTS idx_order_deliveries_plugin ON order_deliveries(fulfillment_plugin_id);
CREATE INDEX IF NOT EXISTS idx_order_deliveries_status ON order_deliveries(status);

CREATE TABLE IF NOT EXISTS order_delivery_facts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    fact_id TEXT NOT NULL UNIQUE,
    delivery_no TEXT NOT NULL,
    order_no TEXT NOT NULL DEFAULT '',
    fulfillment_plugin_id TEXT NOT NULL DEFAULT '',
    fact_type TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT '',
    result_payload_json TEXT NOT NULL DEFAULT '',
    idempotency_key TEXT NOT NULL DEFAULT '',
    retryable INTEGER NOT NULL DEFAULT 0,
    error_code TEXT NOT NULL DEFAULT '',
    error_message TEXT NOT NULL DEFAULT '',
    occurred_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_order_delivery_facts_delivery_no ON order_delivery_facts(delivery_no);
CREATE INDEX IF NOT EXISTS idx_order_delivery_facts_order_no ON order_delivery_facts(order_no);
CREATE INDEX IF NOT EXISTS idx_order_delivery_facts_plugin ON order_delivery_facts(fulfillment_plugin_id);

CREATE TABLE IF NOT EXISTS order_state_events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    order_id INTEGER NOT NULL,
    order_no TEXT NOT NULL DEFAULT '',
    event_type TEXT NOT NULL DEFAULT '',
    from_status TEXT NOT NULL DEFAULT '',
    to_status TEXT NOT NULL DEFAULT '',
    payment_status TEXT NOT NULL DEFAULT '',
    delivery_status TEXT NOT NULL DEFAULT '',
    actor_subject_id TEXT NOT NULL DEFAULT '',
    owner_plugin_id TEXT NOT NULL DEFAULT '',
    payload_json TEXT NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_order_state_events_order_id ON order_state_events(order_id);
CREATE INDEX IF NOT EXISTS idx_order_state_events_order_no ON order_state_events(order_no);
CREATE INDEX IF NOT EXISTS idx_order_state_events_event_type ON order_state_events(event_type);

CREATE TABLE IF NOT EXISTS system_settings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    key TEXT NOT NULL UNIQUE,
    value TEXT NOT NULL DEFAULT '',
    remark TEXT NOT NULL DEFAULT '',
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS email_configs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    enabled INTEGER NOT NULL DEFAULT 0,
    smtp_host TEXT NOT NULL DEFAULT '',
    smtp_port INTEGER NOT NULL DEFAULT 465,
    smtp_user TEXT NOT NULL DEFAULT '',
    smtp_password TEXT NOT NULL DEFAULT '',
    from_name TEXT NOT NULL DEFAULT '',
    from_email TEXT NOT NULL DEFAULT '',
    encryption TEXT NOT NULL DEFAULT 'ssl',
    code_length INTEGER NOT NULL DEFAULT 6,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS system_configs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    system_title TEXT NOT NULL DEFAULT '',
    admin_suffix TEXT NOT NULL DEFAULT '',
    enable_login INTEGER NOT NULL DEFAULT 1,
    admin_human_verification_enabled INTEGER NOT NULL DEFAULT 0,
    admin_human_verification_provider_id TEXT NOT NULL DEFAULT 'keyswift.image_captcha',
    admin_username TEXT NOT NULL DEFAULT '',
    admin_password TEXT NOT NULL DEFAULT '',
    admin_password_initialized INTEGER NOT NULL DEFAULT 0,
    enable2_fa INTEGER NOT NULL DEFAULT 0,
    totp_secret TEXT NOT NULL DEFAULT '',
    enable_session_timeout INTEGER NOT NULL DEFAULT 1,
    session_timeout INTEGER NOT NULL DEFAULT 60,
    user_allow_register INTEGER NOT NULL DEFAULT 1,
    user_login_human_verification_enabled INTEGER NOT NULL DEFAULT 0,
    user_login_human_verification_provider_id TEXT NOT NULL DEFAULT 'keyswift.image_captcha',
    user_register_human_verification_enabled INTEGER NOT NULL DEFAULT 0,
    user_register_human_verification_provider_id TEXT NOT NULL DEFAULT 'keyswift.image_captcha',
    user_register_human_verification_follow_login INTEGER NOT NULL DEFAULT 1,
    user_enable2_fa INTEGER NOT NULL DEFAULT 1,
    user_require_email_verification INTEGER NOT NULL DEFAULT 0,
    user_enable_session_timeout INTEGER NOT NULL DEFAULT 1,
    user_session_timeout INTEGER NOT NULL DEFAULT 120,
    enable_whitelist INTEGER NOT NULL DEFAULT 0,
    ip_whitelist TEXT NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS login_attempts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL DEFAULT '',
    ip TEXT NOT NULL DEFAULT '',
    success INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_login_attempts_username ON login_attempts(username);
CREATE INDEX IF NOT EXISTS idx_login_attempts_ip ON login_attempts(ip);

CREATE TABLE IF NOT EXISTS user_sessions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT NOT NULL UNIQUE,
    user_id INTEGER NOT NULL DEFAULT 0,
    username TEXT NOT NULL DEFAULT '',
    ip TEXT NOT NULL DEFAULT '',
    user_agent TEXT NOT NULL DEFAULT '',
    expires_at DATETIME NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_user_sessions_expires_at ON user_sessions(expires_at);

CREATE TABLE IF NOT EXISTS admin_sessions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT NOT NULL UNIQUE,
    username TEXT NOT NULL DEFAULT '',
    role TEXT NOT NULL DEFAULT '',
    ip TEXT NOT NULL DEFAULT '',
    user_agent TEXT NOT NULL DEFAULT '',
    verified INTEGER NOT NULL DEFAULT 0,
    expires_at DATETIME NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_admin_sessions_expires_at ON admin_sessions(expires_at);

CREATE TABLE IF NOT EXISTS login_failure_records (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    key TEXT NOT NULL UNIQUE,
    failure_count INTEGER NOT NULL DEFAULT 0,
    first_fail_at DATETIME NOT NULL,
    locked_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS manual_kamis (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    product_id INTEGER NOT NULL DEFAULT 0,
    kami_code TEXT NOT NULL DEFAULT '',
    status INTEGER NOT NULL DEFAULT 0,
    order_id INTEGER NOT NULL DEFAULT 0,
    order_no TEXT NOT NULL DEFAULT '',
    sold_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME
);

CREATE INDEX IF NOT EXISTS idx_manual_kamis_product_id ON manual_kamis(product_id);
CREATE INDEX IF NOT EXISTS idx_manual_kamis_deleted_at ON manual_kamis(deleted_at);

CREATE TABLE IF NOT EXISTS admin_roles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL DEFAULT '',
    permissions TEXT NOT NULL DEFAULT '',
    is_system INTEGER NOT NULL DEFAULT 0,
    status INTEGER NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS permission_definitions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    permission_code TEXT NOT NULL UNIQUE,
    owner_type TEXT NOT NULL DEFAULT 'host',
    owner_plugin_id TEXT NOT NULL DEFAULT '',
    risk_level TEXT NOT NULL DEFAULT 'normal',
    group_key TEXT NOT NULL DEFAULT '',
    name TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    default_grant_policy TEXT NOT NULL DEFAULT 'manual',
    status TEXT NOT NULL DEFAULT 'active',
    extensions_json TEXT NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_permission_definitions_owner ON permission_definitions(owner_type, owner_plugin_id);
CREATE INDEX IF NOT EXISTS idx_permission_definitions_group_key ON permission_definitions(group_key);
CREATE INDEX IF NOT EXISTS idx_permission_definitions_status ON permission_definitions(status);

CREATE TABLE IF NOT EXISTS role_permission_grants (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    role_id INTEGER NOT NULL,
    permission_code TEXT NOT NULL,
    granted_by_subject_id TEXT NOT NULL DEFAULT '',
    granted_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(role_id, permission_code)
);

CREATE INDEX IF NOT EXISTS idx_role_permission_grants_role_id ON role_permission_grants(role_id);
CREATE INDEX IF NOT EXISTS idx_role_permission_grants_permission ON role_permission_grants(permission_code);

CREATE TABLE IF NOT EXISTS subject_permission_grants (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    subject_id TEXT NOT NULL,
    subject_type TEXT NOT NULL DEFAULT '',
    permission_code TEXT NOT NULL,
    granted_by_subject_id TEXT NOT NULL DEFAULT '',
    granted_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at DATETIME,
    UNIQUE(subject_id, permission_code)
);

CREATE INDEX IF NOT EXISTS idx_subject_permission_grants_subject ON subject_permission_grants(subject_id, subject_type);
CREATE INDEX IF NOT EXISTS idx_subject_permission_grants_permission ON subject_permission_grants(permission_code);

CREATE TABLE IF NOT EXISTS resource_scope_definitions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    resource_type TEXT NOT NULL,
    scope_type TEXT NOT NULL,
    owner_plugin_id TEXT NOT NULL DEFAULT '',
    name TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'active',
    extensions_json TEXT NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(resource_type, scope_type, owner_plugin_id)
);

CREATE INDEX IF NOT EXISTS idx_resource_scope_definitions_resource ON resource_scope_definitions(resource_type);
CREATE INDEX IF NOT EXISTS idx_resource_scope_definitions_owner ON resource_scope_definitions(owner_plugin_id);

CREATE TABLE IF NOT EXISTS subject_data_scope_grants (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    subject_id TEXT NOT NULL,
    subject_type TEXT NOT NULL DEFAULT '',
    resource_type TEXT NOT NULL,
    scope_type TEXT NOT NULL,
    scope_value TEXT NOT NULL DEFAULT '',
    owner_plugin_id TEXT NOT NULL DEFAULT '',
    granted_by_subject_id TEXT NOT NULL DEFAULT '',
    granted_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at DATETIME,
    UNIQUE(subject_id, resource_type, scope_type, scope_value, owner_plugin_id)
);

CREATE INDEX IF NOT EXISTS idx_subject_data_scope_grants_subject ON subject_data_scope_grants(subject_id, subject_type);
CREATE INDEX IF NOT EXISTS idx_subject_data_scope_grants_resource ON subject_data_scope_grants(resource_type, scope_type);

CREATE TABLE IF NOT EXISTS admins (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL DEFAULT '',
    role_id INTEGER NOT NULL DEFAULT 0,
    email TEXT NOT NULL DEFAULT '',
    nickname TEXT NOT NULL DEFAULT '',
    avatar TEXT NOT NULL DEFAULT '',
    enable2_fa INTEGER NOT NULL DEFAULT 0,
    totp_secret TEXT NOT NULL DEFAULT '',
    status INTEGER NOT NULL DEFAULT 1,
    last_login_at DATETIME,
    last_login_ip TEXT NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_admins_role_id ON admins(role_id);

CREATE TABLE IF NOT EXISTS user_balances (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL UNIQUE,
    balance NUMERIC NOT NULL DEFAULT 0,
    frozen NUMERIC NOT NULL DEFAULT 0,
    total_in NUMERIC NOT NULL DEFAULT 0,
    total_out NUMERIC NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS balance_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL DEFAULT 0,
    type TEXT NOT NULL DEFAULT '',
    amount NUMERIC NOT NULL DEFAULT 0,
    before_balance NUMERIC NOT NULL DEFAULT 0,
    after_balance NUMERIC NOT NULL DEFAULT 0,
    order_no TEXT NOT NULL DEFAULT '',
    remark TEXT NOT NULL DEFAULT '',
    operator_id INTEGER NOT NULL DEFAULT 0,
    operator_type TEXT NOT NULL DEFAULT 'user',
    client_ip TEXT NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_balance_logs_user_id ON balance_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_balance_logs_type ON balance_logs(type);
CREATE INDEX IF NOT EXISTS idx_balance_logs_order_no ON balance_logs(order_no);

CREATE TABLE IF NOT EXISTS product_images (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    product_id INTEGER NOT NULL DEFAULT 0,
    url TEXT NOT NULL DEFAULT '',
    sort_order INTEGER NOT NULL DEFAULT 0,
    is_primary INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_product_images_product_id ON product_images(product_id);

CREATE TABLE IF NOT EXISTS plugin_registry (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    plugin_id TEXT NOT NULL UNIQUE,
    install_id TEXT NOT NULL DEFAULT '',
    current_version TEXT NOT NULL DEFAULT '',
    install_root TEXT NOT NULL DEFAULT '',
    source_type TEXT NOT NULL DEFAULT '',
    source_uri TEXT NOT NULL DEFAULT '',
    enabled INTEGER NOT NULL DEFAULT 0,
    autostart INTEGER NOT NULL DEFAULT 0,
    desired_state TEXT NOT NULL DEFAULT 'approved-disabled',
    actual_state TEXT NOT NULL DEFAULT 'stopped',
    lifecycle_state TEXT NOT NULL DEFAULT 'discovered',
    trust_level TEXT NOT NULL DEFAULT 'local-approved',
    signature_status TEXT NOT NULL DEFAULT 'unknown',
    approved_manifest_hash TEXT NOT NULL DEFAULT '',
    approved_package_hash TEXT NOT NULL DEFAULT '',
    approved_binary_hash TEXT NOT NULL DEFAULT '',
    current_manifest_hash TEXT NOT NULL DEFAULT '',
    last_verified_at DATETIME,
    last_verify_status TEXT NOT NULL DEFAULT '',
    tamper_status TEXT NOT NULL DEFAULT '',
    quarantine_reason TEXT NOT NULL DEFAULT '',
    config_version INTEGER NOT NULL DEFAULT 1,
    selected_os TEXT NOT NULL DEFAULT '',
    selected_arch TEXT NOT NULL DEFAULT '',
    health_status TEXT NOT NULL DEFAULT 'stopped',
    last_start_at DATETIME,
    last_ready_at DATETIME,
    last_stop_at DATETIME,
    last_fault_at DATETIME,
    manifest_json TEXT NOT NULL DEFAULT '',
    extensions_json TEXT NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS plugin_versions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    plugin_id TEXT NOT NULL,
    version TEXT NOT NULL,
    manifest_hash TEXT NOT NULL DEFAULT '',
    package_hash TEXT NOT NULL DEFAULT '',
    binary_hash TEXT NOT NULL DEFAULT '',
    install_path TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'installed',
    installed_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(plugin_id, version)
);

CREATE INDEX IF NOT EXISTS idx_plugin_versions_plugin_id ON plugin_versions(plugin_id);
CREATE INDEX IF NOT EXISTS idx_plugin_versions_status ON plugin_versions(status);

CREATE TABLE IF NOT EXISTS plugin_runtime_sessions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    plugin_id TEXT NOT NULL,
    version TEXT NOT NULL DEFAULT '',
    instance_id TEXT NOT NULL UNIQUE,
    pid INTEGER NOT NULL DEFAULT 0,
    state TEXT NOT NULL DEFAULT 'starting',
    started_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ready_at DATETIME,
    stopped_at DATETIME,
    last_heartbeat_at DATETIME,
    fault_reason TEXT NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_plugin_runtime_sessions_plugin_id ON plugin_runtime_sessions(plugin_id);
CREATE INDEX IF NOT EXISTS idx_plugin_runtime_sessions_state ON plugin_runtime_sessions(state);

CREATE TABLE IF NOT EXISTS plugin_trust_records (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    plugin_id TEXT NOT NULL,
    version TEXT NOT NULL DEFAULT '',
    trust_level TEXT NOT NULL DEFAULT 'local-approved',
    signature_status TEXT NOT NULL DEFAULT 'unknown',
    approved_by TEXT NOT NULL DEFAULT '',
    approved_at DATETIME,
    risk_summary TEXT NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(plugin_id, version)
);

CREATE INDEX IF NOT EXISTS idx_plugin_trust_records_plugin_id ON plugin_trust_records(plugin_id);
CREATE INDEX IF NOT EXISTS idx_plugin_trust_records_trust ON plugin_trust_records(trust_level);

CREATE TABLE IF NOT EXISTS plugin_config_schemas (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    plugin_id TEXT NOT NULL,
    config_key TEXT NOT NULL DEFAULT 'default',
    schema_version TEXT NOT NULL DEFAULT '',
    schema_json TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'active',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(plugin_id, config_key)
);

CREATE INDEX IF NOT EXISTS idx_plugin_config_schemas_plugin_id ON plugin_config_schemas(plugin_id);

CREATE TABLE IF NOT EXISTS plugin_config_values (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    plugin_id TEXT NOT NULL,
    config_key TEXT NOT NULL DEFAULT 'default',
    value_json TEXT NOT NULL DEFAULT '',
    secret_json TEXT NOT NULL DEFAULT '',
    revision INTEGER NOT NULL DEFAULT 1,
    updated_by TEXT NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(plugin_id, config_key)
);

CREATE INDEX IF NOT EXISTS idx_plugin_config_values_plugin_id ON plugin_config_values(plugin_id);

CREATE TABLE IF NOT EXISTS plugin_config_revisions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    plugin_id TEXT NOT NULL,
    config_key TEXT NOT NULL DEFAULT 'default',
    revision INTEGER NOT NULL,
    value_digest TEXT NOT NULL DEFAULT '',
    secret_json TEXT NOT NULL DEFAULT '',
    updated_by TEXT NOT NULL DEFAULT '',
    change_summary TEXT NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(plugin_id, config_key, revision)
);

CREATE INDEX IF NOT EXISTS idx_plugin_config_revisions_plugin_id ON plugin_config_revisions(plugin_id);

CREATE TABLE IF NOT EXISTS plugin_state_events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    plugin_id TEXT NOT NULL,
    from_state TEXT NOT NULL DEFAULT '',
    to_state TEXT NOT NULL DEFAULT '',
    event_type TEXT NOT NULL DEFAULT '',
    reason TEXT NOT NULL DEFAULT '',
    operator_subject_id TEXT NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_plugin_state_events_plugin_id ON plugin_state_events(plugin_id);
CREATE INDEX IF NOT EXISTS idx_plugin_state_events_event_type ON plugin_state_events(event_type);

CREATE TABLE IF NOT EXISTS plugin_fault_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    plugin_id TEXT NOT NULL,
    instance_id TEXT NOT NULL DEFAULT '',
    fault_type TEXT NOT NULL DEFAULT '',
    fault_reason TEXT NOT NULL DEFAULT '',
    stack_trace TEXT NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_plugin_fault_logs_plugin_id ON plugin_fault_logs(plugin_id);
CREATE INDEX IF NOT EXISTS idx_plugin_fault_logs_instance_id ON plugin_fault_logs(instance_id);

CREATE TABLE IF NOT EXISTS plugin_artifacts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    plugin_id TEXT NOT NULL DEFAULT '',
    artifact_id TEXT NOT NULL DEFAULT '',
    relative_path TEXT NOT NULL DEFAULT '',
    artifact_type TEXT NOT NULL DEFAULT '',
    platform TEXT NOT NULL DEFAULT '',
    arch TEXT NOT NULL DEFAULT '',
    size_bytes INTEGER NOT NULL DEFAULT 0,
    hash_algorithm TEXT NOT NULL DEFAULT '',
    hash_value TEXT NOT NULL DEFAULT '',
    is_executable INTEGER NOT NULL DEFAULT 0,
    is_required INTEGER NOT NULL DEFAULT 0,
    group_name TEXT NOT NULL DEFAULT '',
    approved_at DATETIME,
    extensions_json TEXT NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(plugin_id, artifact_id)
);

CREATE INDEX IF NOT EXISTS idx_plugin_artifacts_plugin_id ON plugin_artifacts(plugin_id);

CREATE TABLE IF NOT EXISTS plugin_bindings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    plugin_id TEXT NOT NULL DEFAULT '',
    binding_type TEXT NOT NULL DEFAULT '',
    binding_key TEXT NOT NULL DEFAULT '',
    target_scope TEXT NOT NULL DEFAULT '',
    mount_area TEXT NOT NULL DEFAULT '',
    route_or_view_id TEXT NOT NULL DEFAULT '',
    enabled INTEGER NOT NULL DEFAULT 1,
    order_hint INTEGER NOT NULL DEFAULT 0,
    permission_guard TEXT NOT NULL DEFAULT '',
    extensions_json TEXT NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(plugin_id, binding_type, binding_key)
);

CREATE INDEX IF NOT EXISTS idx_plugin_bindings_plugin_id ON plugin_bindings(plugin_id);
CREATE INDEX IF NOT EXISTS idx_plugin_bindings_binding_type ON plugin_bindings(binding_type);

CREATE TABLE IF NOT EXISTS plugin_event_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    event_id TEXT NOT NULL UNIQUE,
    event_type TEXT NOT NULL DEFAULT '',
    event_version TEXT NOT NULL DEFAULT '',
    occurred_at DATETIME NOT NULL,
    producer_type TEXT NOT NULL DEFAULT '',
    producer_id TEXT NOT NULL DEFAULT '',
    request_id TEXT NOT NULL DEFAULT '',
    subject_id TEXT NOT NULL DEFAULT '',
    resource_type TEXT NOT NULL DEFAULT '',
    resource_id TEXT NOT NULL DEFAULT '',
    payload_json TEXT NOT NULL DEFAULT '',
    extensions_json TEXT NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_plugin_event_logs_event_type ON plugin_event_logs(event_type);
CREATE INDEX IF NOT EXISTS idx_plugin_event_logs_resource ON plugin_event_logs(resource_type, resource_id);

CREATE TABLE IF NOT EXISTS plugin_jobs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    job_id TEXT NOT NULL UNIQUE,
    owner_plugin_id TEXT NOT NULL DEFAULT '',
    job_type TEXT NOT NULL DEFAULT '',
    resource_type TEXT NOT NULL DEFAULT '',
    resource_id TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'pending',
    priority INTEGER NOT NULL DEFAULT 0,
    run_at DATETIME NOT NULL,
    retry_count INTEGER NOT NULL DEFAULT 0,
    max_retries INTEGER NOT NULL DEFAULT 3,
    last_error TEXT NOT NULL DEFAULT '',
    payload_json TEXT NOT NULL DEFAULT '',
    started_at DATETIME,
    finished_at DATETIME,
    extensions_json TEXT NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_plugin_jobs_owner_plugin_id ON plugin_jobs(owner_plugin_id);
CREATE INDEX IF NOT EXISTS idx_plugin_jobs_status_run_at ON plugin_jobs(status, run_at);

CREATE TABLE IF NOT EXISTS plugin_migrations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    plugin_id TEXT NOT NULL DEFAULT '',
    migration_id TEXT NOT NULL DEFAULT '',
    version TEXT NOT NULL DEFAULT '',
    direction TEXT NOT NULL DEFAULT '',
    path TEXT NOT NULL DEFAULT '',
    checksum TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'declared',
    executed_at DATETIME,
    error_message TEXT NOT NULL DEFAULT '',
    extensions_json TEXT NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(plugin_id, migration_id)
);

CREATE INDEX IF NOT EXISTS idx_plugin_migrations_plugin_id ON plugin_migrations(plugin_id);

CREATE TABLE IF NOT EXISTS plugin_database_declarations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    plugin_id TEXT NOT NULL UNIQUE,
    plugin_version TEXT NOT NULL DEFAULT '',
    namespace TEXT NOT NULL DEFAULT '',
    storage_mode TEXT NOT NULL DEFAULT 'host-main-db',
    table_count INTEGER NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'declared',
    extensions_json TEXT NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_plugin_database_declarations_namespace ON plugin_database_declarations(namespace);
CREATE INDEX IF NOT EXISTS idx_plugin_database_declarations_status ON plugin_database_declarations(status);

CREATE TABLE IF NOT EXISTS plugin_database_tables (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    plugin_id TEXT NOT NULL,
    plugin_version TEXT NOT NULL DEFAULT '',
    namespace TEXT NOT NULL DEFAULT '',
    table_key TEXT NOT NULL,
    physical_table_name TEXT NOT NULL UNIQUE,
    table_kind TEXT NOT NULL,
    schema_version TEXT NOT NULL DEFAULT '',
    schema_checksum TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'declared',
    sensitivity TEXT NOT NULL DEFAULT 'internal',
    create_policy TEXT NOT NULL DEFAULT 'on_enable',
    drop_policy TEXT NOT NULL DEFAULT 'manual_only',
    backup_policy TEXT NOT NULL DEFAULT 'include',
    retention_policy TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    last_applied_at DATETIME,
    extensions_json TEXT NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(plugin_id, table_key)
);

CREATE INDEX IF NOT EXISTS idx_plugin_database_tables_plugin_id ON plugin_database_tables(plugin_id);
CREATE INDEX IF NOT EXISTS idx_plugin_database_tables_status ON plugin_database_tables(status);

CREATE TABLE IF NOT EXISTS plugin_database_columns (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    plugin_id TEXT NOT NULL,
    table_id INTEGER NOT NULL,
    column_key TEXT NOT NULL,
    column_name TEXT NOT NULL,
    db_type TEXT NOT NULL,
    logical_type TEXT NOT NULL DEFAULT '',
    nullable INTEGER NOT NULL DEFAULT 0,
    default_value_json TEXT NOT NULL DEFAULT '',
    primary_key INTEGER NOT NULL DEFAULT 0,
    auto_increment INTEGER NOT NULL DEFAULT 0,
    unique_key INTEGER NOT NULL DEFAULT 0,
    indexed INTEGER NOT NULL DEFAULT 0,
    encrypted INTEGER NOT NULL DEFAULT 0,
    secret INTEGER NOT NULL DEFAULT 0,
    reference_type TEXT NOT NULL DEFAULT '',
    reference_target TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    extensions_json TEXT NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(table_id, column_name)
);

CREATE INDEX IF NOT EXISTS idx_plugin_database_columns_plugin_id ON plugin_database_columns(plugin_id);
CREATE INDEX IF NOT EXISTS idx_plugin_database_columns_table_id ON plugin_database_columns(table_id);

CREATE TABLE IF NOT EXISTS plugin_database_indexes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    plugin_id TEXT NOT NULL,
    table_id INTEGER NOT NULL,
    index_key TEXT NOT NULL,
    index_name TEXT NOT NULL,
    columns_json TEXT NOT NULL DEFAULT '',
    unique_index INTEGER NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'declared',
    extensions_json TEXT NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(table_id, index_name)
);

CREATE INDEX IF NOT EXISTS idx_plugin_database_indexes_plugin_id ON plugin_database_indexes(plugin_id);
CREATE INDEX IF NOT EXISTS idx_plugin_database_indexes_table_id ON plugin_database_indexes(table_id);

CREATE TABLE IF NOT EXISTS plugin_database_relations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    plugin_id TEXT NOT NULL,
    table_id INTEGER NOT NULL,
    relation_key TEXT NOT NULL,
    local_column TEXT NOT NULL,
    target_resource_type TEXT NOT NULL,
    target_key TEXT NOT NULL,
    relation_type TEXT NOT NULL DEFAULT 'many_to_one',
    required INTEGER NOT NULL DEFAULT 0,
    on_delete_policy TEXT NOT NULL DEFAULT 'restrict',
    extensions_json TEXT NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(table_id, relation_key)
);

CREATE INDEX IF NOT EXISTS idx_plugin_database_relations_plugin_id ON plugin_database_relations(plugin_id);
CREATE INDEX IF NOT EXISTS idx_plugin_database_relations_table_id ON plugin_database_relations(table_id);

CREATE TABLE IF NOT EXISTS plugin_database_operations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    operation_id TEXT NOT NULL UNIQUE,
    plugin_id TEXT NOT NULL,
    plugin_version TEXT NOT NULL DEFAULT '',
    table_key TEXT NOT NULL DEFAULT '',
    operation_type TEXT NOT NULL,
    path TEXT NOT NULL DEFAULT '',
    requires_review INTEGER NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'pending',
    schema_checksum TEXT NOT NULL DEFAULT '',
    executed_by TEXT NOT NULL DEFAULT '',
    execution_ms INTEGER NOT NULL DEFAULT 0,
    error_message TEXT NOT NULL DEFAULT '',
    extensions_json TEXT NOT NULL DEFAULT '',
    started_at DATETIME,
    finished_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_plugin_database_operations_plugin_id ON plugin_database_operations(plugin_id);
CREATE INDEX IF NOT EXISTS idx_plugin_database_operations_status ON plugin_database_operations(status);

CREATE TABLE IF NOT EXISTS event_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    event_id TEXT NOT NULL UNIQUE,
    event_type TEXT NOT NULL DEFAULT '',
    source_type TEXT NOT NULL DEFAULT '',
    source_id TEXT NOT NULL DEFAULT '',
    owner_plugin_id TEXT NOT NULL DEFAULT '',
    payload_json TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'recorded',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_event_logs_event_type ON event_logs(event_type);
CREATE INDEX IF NOT EXISTS idx_event_logs_source ON event_logs(source_type, source_id);
CREATE INDEX IF NOT EXISTS idx_event_logs_owner_plugin_id ON event_logs(owner_plugin_id);

CREATE TABLE IF NOT EXISTS system_jobs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    job_id TEXT NOT NULL UNIQUE,
    job_type TEXT NOT NULL DEFAULT '',
    owner_plugin_id TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'pending',
    run_at DATETIME NOT NULL,
    attempts INTEGER NOT NULL DEFAULT 0,
    max_attempts INTEGER NOT NULL DEFAULT 3,
    payload_json TEXT NOT NULL DEFAULT '',
    last_error TEXT NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_system_jobs_owner_plugin_id ON system_jobs(owner_plugin_id);
CREATE INDEX IF NOT EXISTS idx_system_jobs_status_run_at ON system_jobs(status, run_at);

CREATE TABLE IF NOT EXISTS audit_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    actor_subject_id TEXT NOT NULL DEFAULT '',
    action TEXT NOT NULL DEFAULT '',
    resource_type TEXT NOT NULL DEFAULT '',
    resource_id TEXT NOT NULL DEFAULT '',
    risk_level TEXT NOT NULL DEFAULT 'normal',
    ip TEXT NOT NULL DEFAULT '',
    user_agent TEXT NOT NULL DEFAULT '',
    payload_digest TEXT NOT NULL DEFAULT '',
    payload_json TEXT NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_audit_logs_actor_subject_id ON audit_logs(actor_subject_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_resource ON audit_logs(resource_type, resource_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs(action);
