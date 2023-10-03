/* eslint-disable no-undef */
const ESLintPlugin = require('eslint-webpack-plugin');

module.exports = {
    entry: './src/js/main.js',
    plugins: [new ESLintPlugin()],
    output: {
        path: __dirname + '/dist',
        filename: 'bundle.js'
    },
    module: {
        rules: [
            {
                test: /\.m?ts$|\.tsx?$/,
                // exclude: /node_modules/,
                use: {
                    loader: "ts-loader",
                    options: {
                        onlyCompileBundledFiles: true,
                    }
                },
            },
            {
            test: /.m?js$/,
            exclude: /node_modules/,
            use: {
                loader: 'babel-loader',
                options: {
                    presets: ['@babel/preset-env']
                }
            }
        },
        ]
    }
};