UPDATE sites_config
     SET parser_config = REPLACE(parser_config, '--sites=gs_gs_', '--sites=gs_')
     WHERE parser_config LIKE '%gs_gs_%';
