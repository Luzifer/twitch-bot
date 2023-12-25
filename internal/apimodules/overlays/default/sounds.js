/* global Vue */

import EventClient from './eventclient.js'

new Vue({
  created() {
    this.createAudioInterface()

    new EventClient({
      handlers: {
        custom: ({ fields }) => this.handleCustom(fields),
      },
    })
  },

  data: {
    alerts: [],
    alertsEnabled: true,
    sound: null,
    soundsActive: false,
  },

  el: '#app',

  methods: {
    createAudioInterface() {
      // Create basic audio element to play sound
      this.sound = new Audio()

      this.sound.addEventListener('load', () => this.sound.play(), true)
      this.sound.addEventListener('playing', () => {
        this.soundsActive = true
      })
      this.sound.addEventListener('pause', () => {
        this.soundsActive = false
      })
      this.sound.autoplay = true
      this.sound.crossOrigin = 'anonymous'

      // Create limiter chain and connect audio element to it
      const audioCtx = new AudioContext()

      const preGainNode = audioCtx.createGain()
      // preGainNode.gain.setValueAtTime(-0.5, audioCtx.currentTime) // Decibel

      const masterGainNode = audioCtx.createGain()
      masterGainNode.gain.setValueAtTime(-0.1, audioCtx.currentTime) // Decibel

      const limiterNode = audioCtx.createDynamicsCompressor()
      limiterNode.threshold.setValueAtTime(-30.0, audioCtx.currentTime) // Decibel
      limiterNode.knee.setValueAtTime(0, audioCtx.currentTime) // Decibel
      limiterNode.ratio.setValueAtTime(20.0, audioCtx.currentTime) // Decibel
      limiterNode.attack.setValueAtTime(0, audioCtx.currentTime) // Seconds
      limiterNode.release.setValueAtTime(0.1, audioCtx.currentTime) // Seconds

      preGainNode.connect(limiterNode)
      limiterNode.connect(masterGainNode)
      masterGainNode.connect(audioCtx.destination)

      const source = audioCtx.createMediaElementSource(this.sound)
      source.connect(preGainNode)
    },

    handleCustom(data) {
      switch (data.type) {
      case 'soundalert':
        this.queueAlert({
          soundUrl: data.soundUrl,
        })
        break

      default:
        console.log(`Unhandled custom event ${data.type}`, data)
      }
    },

    playSound(soundUrl) {
      this.soundsActive = true
      this.sound.src = soundUrl
    },

    queueAlert(alertParams) {
      this.alerts.push(alertParams)
    },

    triggerAlert() {
      const alrt = this.alerts.shift()
      if (!alrt) {
        return
      }

      if (!this.alertsEnabled) {
        // Alerts are disabled completely or this one is blocked, discard alert
        return this.triggerAlert()
      }

      if (alrt.soundUrl) {
        this.playSound(alrt.soundUrl)
      }
    },
  },

  name: 'SoundOverlay',

  watch: {
    alerts(to) {
      if (to.length > 0 && !this.alertsRunning) {
        this.triggerAlert()
      }
    },

    soundsActive(to) {
      if (!to) {
        this.triggerAlert()
      }
    },
  },
})
