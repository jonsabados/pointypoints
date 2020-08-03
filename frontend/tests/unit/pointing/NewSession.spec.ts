import { shallowMount } from '@vue/test-utils'
import NewSession from '@/pointing/NewSession.vue'

describe('NewSession', () => {
  it('disables and enables the submit button based on facilitator name', () => {
    const wrapper = shallowMount(NewSession, {
      attachToDocument: true
    })

    expect(wrapper.find('#startSessionButton').is(':disabled')).toBeTruthy()
    wrapper.find('#facilitatorName').setValue('John Doe')
    expect(wrapper.find('#startSessionButton').is(':disabled')).toBeFalsy()
    wrapper.find('#facilitatorName').setValue('')
    expect(wrapper.find('#startSessionButton').is(':disabled')).toBeTruthy()

    wrapper.destroy()
  })
})
