import { createLocalVue, shallowMount } from '@vue/test-utils'
import NewSession from '@/pointing/NewSession.vue'
import Vuex from 'vuex'
import sinon from 'sinon'

describe('NewSession', () => {
  it('clears the current session when mounted', () => {
    const localVue = createLocalVue()
    localVue.use(Vuex)

    const state = {
      pointingSession: {
      }
    }

    const actions = {
      endSession: sinon.spy()
    }

    const store = new Vuex.Store({
      state,
      actions
    })

    const wrapper = shallowMount(NewSession, {
      attachToDocument: true,
      localVue,
      store
    })

    expect(actions.endSession.calledOnce).toBeTruthy()

    wrapper.destroy()
  })

  it('disables and enables the submit button based on facilitator name', () => {
    const localVue = createLocalVue()
    localVue.use(Vuex)

    const state = {
      pointingSession: {
      }
    }

    const actions = {
      endSession: sinon.spy()
    }

    const store = new Vuex.Store({
      state,
      actions
    })

    const wrapper = shallowMount(NewSession, {
      attachToDocument: true,
      localVue,
      store
    })

    expect(wrapper.find('#startSessionButton').is(':disabled')).toBeTruthy()
    wrapper.find('#facilitatorName').setValue('John Doe')
    expect(wrapper.find('#startSessionButton').is(':disabled')).toBeFalsy()
    wrapper.find('#facilitatorName').setValue('')
    expect(wrapper.find('#startSessionButton').is(':disabled')).toBeTruthy()

    wrapper.destroy()
  })
})
