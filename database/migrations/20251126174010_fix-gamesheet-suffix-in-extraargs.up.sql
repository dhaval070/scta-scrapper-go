  UPDATE sites_config
   SET parser_config = JSON_SET(
       parser_config,
       '$.extra_args',
       JSON_ARRAY(CONCAT('--sites=gs_', REPLACE(REPLACE(JSON_UNQUOTE(JSON_EXTRACT(parser_config, '$.extra_args[0]')), '--sites=', ''), '_gs ', ' ')))
   )
   WHERE site_name IN (
       'gs_ehf202425',
       'gs_neghlf202425',
       'gs_neghlw202425',
       'gs_ehffsp202425',
       'gs_sp202526',
       'gs_phlone202526',
       'gs_ehf202526',
       'gs_neghlf202526',
       'gs_neghlw202526',
       'gs_omha202526',
       'gs_sjhshl202526'
   );
