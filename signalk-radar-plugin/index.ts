import { Plugin, ServerAPI } from '@signalk/server-api';
import { IRouter, Request, Response } from 'express';


import * as openapi from './openApi.json';

const CONFIG_SCHEMA = {
  properties: {
    radar: {
      type: 'object',
      title: 'Radar API.',
      description: 'Radarsettings.',
      properties: {
        enable: {
          type: 'boolean',
          default: false,
          title: 'Enable Radar',
          description: ' '
        },
        radarServerUrl: {
          type: 'string',
          title: 'Radar server url',
          default: 'http://localhost:3001',
          description: 'Url of the radar server'
        }
      }
    }
  }
};

const CONFIG_UISCHEMA = {
  radar: {
    enable: {
      'ui:widget': 'checkbox',
      'ui:title': ' ',
      'ui:help': ' '
    },
    radarServerUrl: {
      'ui:disabled': false,
      'ui-help': ''
    },
  }
};

interface RadarConfig {
  enable: boolean;
  radarServerUrl: string;
}

interface SETTINGS {
  radar: RadarConfig;
}

module.exports = (server: ServerAPI): Plugin => {
  let settings: SETTINGS = {
    radar: {
      enable: false,
      radarServerUrl: ''
    }
  };

  const plugin: Plugin = {
    id: 'radar-sk',
    name: 'Radar',
    schema: () => CONFIG_SCHEMA,
    uiSchema: () => CONFIG_UISCHEMA,
    start: (settings: any) => doStartup(settings),
    stop: () => doShutdown(),
    registerWithRouter: (router) =>  doRegisterEndpoints(router),
    getOpenApi: () => openapi
  };


  const doStartup = (options: SETTINGS) => {
    try {
      server.debug(`Starting.`);

      if (typeof options !== 'undefined') {
        settings = options;
      }

      server.debug(`Applied config: ${JSON.stringify(settings)}`);

      let msg = '';
      if (settings.radar.enable) {
        msg = `Started - Providing: radar`;
      }
      server.setPluginStatus(msg);
    } catch (error: any) {
      const msg = 'Started with errors!';
      server.setPluginError(msg);
      server.error('** EXCEPTION: **');
      server.error(error.stack);
      return error;
    }
  };

  
  const doShutdown = () => {
    server.debug('** shutting down **');
    const msg = 'Stopped';
    server.setPluginStatus(msg);
  };

  const doRegisterEndpoints = (router: IRouter) => {
    server.debug(`Initialising endpoints.`);

    const radarPath = '/v1/api/radars';
    //Proxy request to radar server
    router.all(`${radarPath}`, async (req: Request, res: Response) => {      
      if(settings.radar.enable) {
        let options:RequestInit = {
          method:req.method
        }
        if ("POST".indexOf(req.method) >=0) {
          options.body = req.body
        }
        let response = await fetch(`${settings.radar.radarServerUrl}${radarPath}`,options)
        server.debug(`${req.method} ${radarPath}`);
        res.status(response.status);
        if(response.status==200) {
          let json = await response.json();
          res.json(json);            
        } else {
          res.json({});
        }
      } else {
        res.status(404);        
        res.json({});
      }
    });   
    router.get('/settings', (req: Request, res: Response) => {
      res.status(200).json({
        settings: settings
      });
    });
  };
  return plugin;
};
