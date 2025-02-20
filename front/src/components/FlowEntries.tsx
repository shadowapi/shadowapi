import {
  Item,
  ListView,
  DragAndDropOptions,
  Heading,
} from '@adobe/react-spectrum'
import type { ListData } from '@adobe/react-spectrum'
import DragHandle from '@spectrum-icons/workflow/DragHandle';

interface Item {
  id: number;
  uuid?: string;
  type?: string;
  title: string;
  parent?: number;
}

export interface FlowEntriesProp {
  list: ListData<Item>;
  dragAndDropOptions?: DragAndDropOptions;
}

export const FlowEntries = (props: FlowEntriesProp) => {
  const { list } = props;
  const renderList: {
    [key: number]: Item & { children: Item[] }
  } = {}
  for (const item of list.items) {
    if (item.parent) {
      renderList[item.parent].children.push(item)
    } else {
      renderList[item.id] = { ...item, children: [] }
    }
  }
  return (
    <>
      {Object.values(renderList).map((item: Item & { children: Item[] }) => (
        <div key={item.id}>
          <Heading level={5}>{item.title}</Heading>
          <ListView
            key={item.id}
            density="compact"
            items={item.children}
            selectionMode="none"
            aria-label={`Flow entry ${item.title}`}
            dragAndDropHooks={props.dragAndDropOptions}
          >
            {(item: Item) => (<Item textValue={item.title}><DragHandle />{item.title}</Item>)}
          </ListView>
        </div>
      ))}
    </>
  );
}
