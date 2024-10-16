BEGIN;

ALTER TYPE net_error ADD VALUE 'devp2p_handshake_eof';
ALTER TYPE net_error ADD VALUE 'devp2p_disconnect_requested';
ALTER TYPE net_error ADD VALUE 'devp2p_network_error';
ALTER TYPE net_error ADD VALUE 'devp2p_breach_of_protocol';
ALTER TYPE net_error ADD VALUE 'devp2p_useless_peer';
ALTER TYPE net_error ADD VALUE 'devp2p_too_many_peers';
ALTER TYPE net_error ADD VALUE 'devp2p_already_connected';
ALTER TYPE net_error ADD VALUE 'devp2p_incompatible_p2p_protocol_version';
ALTER TYPE net_error ADD VALUE 'devp2p_invalid_node_identity';
ALTER TYPE net_error ADD VALUE 'devp2p_client_quitting';
ALTER TYPE net_error ADD VALUE 'devp2p_unexpected_identity';
ALTER TYPE net_error ADD VALUE 'devp2p_connected_to_self';
ALTER TYPE net_error ADD VALUE 'devp2p_read_timeout';
ALTER TYPE net_error ADD VALUE 'devp2p_subprotocol_error';
ALTER TYPE net_error ADD VALUE 'devp2p_ethprotocol_error';

COMMIT