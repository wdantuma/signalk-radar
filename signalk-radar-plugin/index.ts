import { Plugin, ServerAPI } from '@signalk/server-api';
import { IRouter, Application, Request, Response, Router } from 'express';
import { Radar } from '../radar-client/src/service/radar/radar.model'



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

export interface ParsedResponse {
  [key: string]: Radar;
}

let radarData: Map<string, Radar> = new Map<string, Radar>();
let server: RadarPlugin;
let pluginId: string;


interface SETTINGS {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  radar: RadarConfig;
}

export interface RadarPlugin
  extends Application,
  Omit<ServerAPI, 'registerPutHandler'> {
  config: {
    ssl: boolean;
    configPath: string;
    version: string;
    getExternalPort: () => number;
  };

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  //handleMessage: (id: string | null, msg: any, version?: string) => void;
  streambundle: {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    getSelfBus: (path: string | void) => any;
  };
  registerPutHandler: (
    context: string,
    path: string,
    callback: (
      context: string,
      path: string,
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      value: any,
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      actionResultCallback: (actionResult: any) => void
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
    ) => any
  ) => void;
}

module.exports = (server: RadarPlugin): Plugin => {
  // ** default configuration settings
  let settings: SETTINGS = {
    radar: {
      enable: false,
      radarServerUrl: ''
    }
  };

  // ******** REQUIRED PLUGIN DEFINITION *******
  const plugin: Plugin = {
    id: 'radar-sk',
    name: 'Radar',
    schema: () => CONFIG_SCHEMA,
    uiSchema: () => CONFIG_UISCHEMA,
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    start: (settings: any) => {
      doStartup(settings);
    },
    stop: () => {
      doShutdown();
    },
    registerWithRouter: (router) => {
      return initApiEndpoints(router);
    },
    getOpenApi: () => openapi
  };
  // ************************************

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
        fetchRadarData(settings.radar);
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

  const fetchRadarData = async (config: RadarConfig) => {
    server.debug("Fetching data.");
    let response = await fetch(`${config.radarServerUrl}/v1/api/radars`)
    radarData = await response.json()
  };

  const doShutdown = () => {
    server.debug('** shutting down **');
    radarData = new Map<string, Radar>();
    const msg = 'Stopped';
    server.setPluginStatus(msg);
  };

  const initApiEndpoints = (router: IRouter) => {
    server.debug(`Initialising endpoints.`);

    const radarPath = '/v1/api/radars';
    router.get(`${radarPath}`, async (req: Request, res: Response) => {
      server.debug(`${req.method} ${radarPath}`);
      res.status(200);
      res.json(radarData);
    });
    router.get(`${radarPath}/:id`, async (req: Request, res: Response) => {
      server.debug(`${req.method} ${radarPath}/:id`);
      const r =
        radarData && radarData.get(req.params.id)
          ? radarData.get(req.params.id)
          : {};
      res.status(200);
      res.json(r);
    });
    router.get('/settings', (req: Request, res: Response) => {
      res.status(200).json({
        settings: settings
      });
    });
  };
  return plugin;
};
