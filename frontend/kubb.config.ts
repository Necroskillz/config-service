import { defineConfig } from '@kubb/core';
import { pluginOas } from '@kubb/plugin-oas';
import { pluginReactQuery } from '@kubb/plugin-react-query';
import { pluginTs } from '@kubb/plugin-ts';

export default defineConfig(() => {
  return {
    root: '.',
    input: {
      path: '../backend/docs/swagger.yaml',
    },
    output: {
      path: './src/gen',
    },
    plugins: [pluginTs(), pluginOas(), pluginReactQuery({ client: { importPath: '~/axios' } })],
  };
});
