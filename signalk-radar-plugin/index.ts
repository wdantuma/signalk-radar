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
    radarServerUrl: {
      'ui:disabled': false,
      'ui-help': ''
    },
  }
};

interface RadarConfig {
  enable: boolean,
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
        settings = {...settings,...options};
      }
      settings.radar.enable = true

      server.debug(`Applied config: ${JSON.stringify(settings)}`);

      server.setPluginStatus(`Started - Providing: radar`);
    } catch (error: any) {
      server.setPluginError('Started with errors!');
      server.error('** EXCEPTION: **');
      server.error(error.stack);
      return error;
    }
  };

  
  const doShutdown = () => {
    server.debug('** shutting down **');
    settings.radar.enable = false
    const msg = 'Stopped';
    server.setPluginStatus(msg);
  };

  const doRegisterEndpoints = (router: IRouter) => {
    server.debug(`Initialising endpoints.`);

    const radarPath = '/v1/api/radars';
    //Proxy request to radar server
    router.all(`${radarPath}`, async (req: Request, res: Response) => {     
      if(settings.radar.enable) {
        try {
          let options:RequestInit = {
            method:req.method
          }
          if ("POST".indexOf(req.method) >=0) {
            options.body = req.body
          }
          let response = await fetch(`${settings.radar.radarServerUrl}${radarPath}`,options)
          server.debug(`${req.method} ${radarPath}`);
          if(response.status==200) {
            let json = await response.json();
            res.status(response.status).json(json);            
          } else {
            res.status(response.status).send(response.body);  
          }  
        } catch {
          res.status(504).send("");
        }
      } else {     
        server.debug(`${req.method} ${radarPath}`);   
        res.status(404).send("");
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
