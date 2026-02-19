# relm4 Component Patterns

## Full AsyncComponent Template

```rust
use relm4::prelude::*;

pub struct MyModel {
    loading: bool,
    items: Vec<String>,
    error: Option<String>,
}

#[derive(Debug)]
pub enum MyMsg {
    Load,
    SetItems(Vec<String>),
    SetError(String),
    RemoveItem(usize),
}

#[derive(Debug)]
pub enum MyCmd {
    FetchItems,
}

#[relm4::component(async, pub)]
impl AsyncComponent for MyModel {
    type Init = ();
    type Input = MyMsg;
    type Output = ();
    type CommandOutput = MyCmd;

    view! {
        adw::ApplicationWindow {
            set_title: Some("My App"),
            set_default_width: 800,
            set_default_height: 600,

            adw::ToolbarView {
                add_top_bar = &adw::HeaderBar {
                    #[wrap(Some)]
                    set_title_widget = &adw::WindowTitle {
                        set_title: "My App",
                    }
                },

                #[wrap(Some)]
                set_content = &gtk::Box {
                    set_orientation: gtk::Orientation::Vertical,

                    if model.loading {
                        gtk::Spinner {
                            set_spinning: true,
                            set_halign: gtk::Align::Center,
                            set_valign: gtk::Align::Center,
                        }
                    } else if model.items.is_empty() {
                        adw::StatusPage {
                            set_icon_name: Some("list-add-symbolic"),
                            set_title: "No Items",
                            set_description: Some("Click Load to fetch items"),
                        }
                    } else {
                        gtk::ScrolledWindow {
                            set_hexpand: true,
                            set_vexpand: true,

                            gtk::ListBox {
                                set_selection_mode: gtk::SelectionMode::None,
                            }
                        }
                    }
                }
            }
        }
    }

    async fn init(
        _init: Self::Init,
        root: Self::Root,
        sender: AsyncComponentSender<Self>,
    ) -> AsyncComponentParts<Self> {
        let model = MyModel {
            loading: false,
            items: vec![],
            error: None,
        };
        let widgets = view_output!();
        AsyncComponentParts { model, widgets }
    }

    async fn update(
        &mut self,
        msg: Self::Input,
        sender: AsyncComponentSender<Self>,
        _root: &Self::Root,
    ) {
        match msg {
            MyMsg::Load => {
                self.loading = true;
                self.error = None;
                sender.command(|_, _| MyCmd::FetchItems);
            }
            MyMsg::SetItems(items) => {
                self.loading = false;
                self.items = items;
            }
            MyMsg::SetError(e) => {
                self.loading = false;
                self.error = Some(e);
            }
            MyMsg::RemoveItem(idx) => {
                self.items.remove(idx);
            }
        }
    }

    async fn update_cmd(
        &mut self,
        cmd: Self::CommandOutput,
        sender: AsyncComponentSender<Self>,
        _root: &Self::Root,
    ) {
        match cmd {
            MyCmd::FetchItems => {
                match fetch_items().await {
                    Ok(items) => sender.input(MyMsg::SetItems(items)),
                    Err(e)    => sender.input(MyMsg::SetError(e.to_string())),
                }
            }
        }
    }
}
```

---

## Factory Pattern for Dynamic Lists

Use `FactoryVecDeque` for dynamic lists where items have their own update logic:

```rust
use relm4::factory::FactoryVecDeque;

pub struct ListModel {
    items: FactoryVecDeque<ItemModel>,
}

#[derive(Debug)]
pub enum ListMsg {
    AddItem(String),
    RemoveItem(usize),
}

#[relm4::component(pub)]
impl SimpleComponent for ListModel {
    type Init = ();
    type Input = ListMsg;
    type Output = ();

    view! {
        gtk::Box {
            set_orientation: gtk::Orientation::Vertical,

            gtk::ScrolledWindow {
                set_vexpand: true,

                #[local_ref]
                items_box -> gtk::ListBox {
                    set_selection_mode: gtk::SelectionMode::None,
                }
            }
        }
    }

    fn init(
        _init: Self::Init,
        root: Self::Root,
        sender: ComponentSender<Self>,
    ) -> ComponentParts<Self> {
        let items = FactoryVecDeque::builder()
            .launch(gtk::ListBox::default())
            .detach();

        let model = ListModel { items };
        let items_box = model.items.widget();
        let widgets = view_output!();
        ComponentParts { model, widgets }
    }

    fn update(&mut self, msg: Self::Input, _sender: ComponentSender<Self>) {
        let mut guard = self.items.guard();
        match msg {
            ListMsg::AddItem(text) => { guard.push_back(text); }
            ListMsg::RemoveItem(idx) => { guard.remove(idx); }
        }
    }
}
```

---

## Parent–Child Communication

```
Parent                     Child
  ↓ sender.input()          ↑ sender.output()
AppMsg::ChildAction  ←←←  ChildMsg::Done
```

```rust
// In parent — connect child output to parent input
let child = ChildModel::builder()
    .launch(init)
    .forward(sender.input_sender(), |out| match out {
        ChildOutput::Done(data) => AppMsg::ChildDone(data),
        ChildOutput::Cancel     => AppMsg::ChildCancelled,
    });
```

---

## Conditional View Rendering

relm4's `view!` macro supports `if`/`else` and `match` branches:

```rust
view! {
    gtk::Stack {
        // Show spinner when loading
        if model.loading {
            gtk::Spinner {
                set_spinning: true,
            }
        } else if model.items.is_empty() {
            adw::StatusPage {
                set_title: "Empty",
            }
        } else {
            gtk::ScrolledWindow {
                // content
            }
        }
    }
}
```

---

## Accessing Named Widgets Outside view!

```rust
view! {
    gtk::Box {
        #[name = "entry"]
        gtk::Entry {
            set_placeholder_text: Some("Enter URL"),
        }
    }
}

// In init():
widgets.entry.grab_focus();
widgets.entry.connect_activate(move |e| {
    sender.input(Msg::Submit(e.text().to_string()));
});
```

---

## CSS Styling

```rust
// Load CSS in application startup
let provider = gtk::CssProvider::new();
provider.load_from_string(
    ".error-entry { border-color: @error_color; }"
);
gtk::style_context_add_provider_for_display(
    &gtk::gdk::Display::default().unwrap(),
    &provider,
    gtk::STYLE_PROVIDER_PRIORITY_APPLICATION,
);

// Apply class dynamically
if !valid {
    entry.add_css_class("error-entry");
} else {
    entry.remove_css_class("error-entry");
}
```
