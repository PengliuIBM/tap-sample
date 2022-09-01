<template>
  <div></div>
</template>

<script>
/* eslint-disable */

/*! Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
*  SPDX-License-Identifier: MIT-0
*/

import { bus } from '../main' 
const mqtt = require('mqtt') 
const topics = {
  //publish: 'iot-topic',
  publish: 'alleycat-subscribe',
  subscribe: 'rtinfo'
}

export default {
  name: 'IoT',
  mounted: async function () {
    const AWSConfiguration = this.$appConfig
    console.log('IoT mounted')

    const clientId = 'alleycat-' + (Math.floor((Math.random() * 100000) + 1))

    const that = this
    const mqttserver = 'rmq.haas-495.pez.vmware.com'
    const mqttconfig = {
       path: '/ws',
       username: 'jeffrey',
       password: '@abc12345D'
    }

    const mqttClient  = mqtt.connect("ws://rmq.haas-495.pez.vmware.com:15675", mqttconfig);
    // When first connected, subscribe to the topics we are interested in.
    mqttClient.on('connect', function () {
      console.log('mqttClient connected')
      mqttClient.subscribe(topics.subscribe)
    })
    // Attempt to reconnect in the event of any error
    mqttClient.on('error', async function (err) {
      console.log('mqttClient error:', err)
    })

    // Publish message to IoT Core topic
    bus.$on('publish', async (data) => {
      console.log('Publish: ', data)
      mqttClient.publish(topics.publish, JSON.stringify(data))
    })

    // A message has arrived - parse to determine topic
    mqttClient.on('message', function (topic, payload) {
	    //console.log('message ',payload) 
      const payloadEnvelope = JSON.parse(payload.toString())
      console.log('IoT::onMessage: ', topic, payloadEnvelope)
      bus.$emit('message', payloadEnvelope)
    })
  }
}
</script>