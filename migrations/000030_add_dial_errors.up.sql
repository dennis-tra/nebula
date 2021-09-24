ALTER TYPE dial_error ADD VALUE 'stream_reset'; -- getting closest peer with CPL 2: stream reset
ALTER TYPE dial_error ADD VALUE 'host_is_down';
ALTER TYPE dial_error ADD VALUE 'negotiate_security_protocol_no_trailing_new_line'; -- failed to negotiate security protocol: message did not have trailing newline
